package models

import (
	"database/sql"
	"fmt"
	"itsky/a2b-api-go/utils"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func GetCard(username string) *Card {
	result := db.QueryRow("SELECT useralias,credit FROM cc_card WHERE useralias = ?", username)

	client := Card{}
	err := result.Scan(&client.Useralias, &client.Credit)
	if err != nil {
		utils.Log.Println(err)
		return nil
	}
	return &client
}

// https://github.com/Star2Billing/a2billing/blob/develop/admin/Public/form_data/FG_var_payment.inc
// https://github.com/Star2Billing/a2billing/blob/8dd474c6077544dcc757159a50149bbeb403c314/common/lib/Form/Class.FormHandler.inc.php#L1457
// https://github.com/Star2Billing/a2billing/blob/8dd474c6077544dcc757159a50149bbeb403c314/common/lib/Form/Class.FormBO.php#L884
func CardRecharge(useralias string, amount int, paymentTxRef string, paymentDate time.Time) (*Card, error) {
	// use transaction
	tx, err := db.Begin()
	if err != nil {
		utils.Log.Println(err)
		return nil, err
	}
	defer tx.Rollback()

	card := Card{}
	if err := tx.QueryRow(`
SELECT id,useralias,credit,vat
FROM cc_card WHERE useralias = ?`,
		useralias).Scan(&card.ID, &card.Useralias, &card.Credit, &card.Vat); err != nil {
		utils.Log.Println(err)
		return nil, nil
	}

	paymentType := 0
	createPayment, err := tx.Exec(`
INSERT INTO cc_logpayment (date, payment, card_id, description, added_refill, payment_type)
VALUES (?, ?, ?, ?, ?, ?)
		`, paymentDate, amount, card.ID, "Recharge API "+paymentTxRef, 1, paymentType)
	if err != nil {
		utils.Log.Println(err)
		return nil, err
	}
	paymentId, _ := createPayment.LastInsertId()

	amountWithoutVat := float64(amount) / (1 + (card.Vat / 100))
	createRefill, err := tx.Exec(`
INSERT INTO cc_logrefill (date, credit, card_id, description, refill_type)
VALUES (?, ?, ?, ?, ?)
	`, paymentDate, amountWithoutVat, card.ID, "Recharge API "+paymentTxRef, paymentType /* refill_type == payment_type */)
	if err != nil {
		utils.Log.Println(err)
		return nil, err
	}
	refillId, _ := createRefill.LastInsertId()

	// update card credit
	card.Credit += amountWithoutVat
	if _, err := tx.Exec("UPDATE cc_card SET credit = ? WHERE id = ?",
		card.Credit, card.ID); err != nil {
		utils.Log.Println(err)
		return nil, err
	}

	// link refill and payment
	if _, err := tx.Exec("UPDATE cc_logpayment SET id_logrefill = ? WHERE id = ?",
		refillId, paymentId); err != nil {
		utils.Log.Println(err)
		return nil, err
	}

	// create associated invoice
	isNewConf := false
	invoiceConf := InvoiceConf{Key: fmt.Sprintf("count_%d", paymentDate.Year())}
	if err := tx.QueryRow("SELECT * FROM cc_invoice_conf WHERE key_val = ?", invoiceConf.Key).
		Scan(&invoiceConf.ID,
			&invoiceConf.Key, &invoiceConf.Value); err != nil {
		if err != sql.ErrNoRows {
			utils.Log.Println(err)
			return nil, err
		}
		isNewConf = true
	}

	count, err := strconv.Atoi(invoiceConf.Value)
	if err != nil {
		count = 0
	}
	count += 1

	if isNewConf {
		if _, err := tx.Exec("INSERT INTO cc_invoice_conf (key_val, value) VALUES (?, ?)",
			invoiceConf.Key, strconv.Itoa(count)); err != nil {
			utils.Log.Println(err)
			return nil, err
		}
	} else {
		if _, err := tx.Exec("UPDATE cc_invoice_conf SET value = ? WHERE key_val = ?",
			strconv.Itoa(count), invoiceConf.Key); err != nil {
			utils.Log.Println(err)
			return nil, err
		}
	}

	description := "Invoice for refill (api)"
	reference := fmt.Sprintf("%d%08d", paymentDate.Year(), count)
	createInvoice, err := tx.Exec(`INSERT INTO cc_invoice (date, id_card, title, reference, description, status, paid_status)
VALUES (?, ?, ?, ?, ?, ?, ?)`, paymentDate, card.ID, "REFILL", reference, description, 1, 1)
	if err != nil {
		utils.Log.Println(err)
		return nil, err
	}
	invoiceId, _ := createInvoice.LastInsertId()

	// add payment to invoice
	if _, err := tx.Exec("INSERT INTO cc_invoice_payment (id_invoice, id_payment) VALUES (?, ?)",
		invoiceId, paymentId); err != nil {
		utils.Log.Println(err)
		return nil, err
	}

	if _, err := tx.Exec("INSERT INTO cc_invoice_item (date, id_invoice, price, vat, description) VALUES (?, ?, ?, ?, ?)",
		paymentDate, invoiceId, amountWithoutVat, card.Vat, description); err != nil {
		utils.Log.Println(err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		utils.Log.Println(err)
		return nil, err
	}
	return &card, nil
}
