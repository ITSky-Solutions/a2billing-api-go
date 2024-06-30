package models

import (
	"database/sql"
	"fmt"
	"itsky/a2b-api-go/env"
	"itsky/a2b-api-go/utils"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func getDB() *sql.DB {
	var connStr string
	if env.Env.DbPassword == "" {
		connStr = fmt.Sprintf("%s@tcp(%s:%s)/%s", env.Env.DbUser, env.Env.DbHost, env.Env.DbPort, env.Env.DbName)
	} else {
		connStr = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			env.Env.DbUser, env.Env.DbPassword, env.Env.DbHost, env.Env.DbPort, env.Env.DbName)
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return db
}

func GetCard(username string) *Card {
	db := getDB()
	defer db.Close()

	result := db.QueryRow("SELECT useralias,credit FROM cc_card WHERE useralias = ?", username)

	client := Card{}
	err := result.Scan(&client.Useralias, &client.Credit)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &client
}

// https://github.com/Star2Billing/a2billing/blob/8dd474c6077544dcc757159a50149bbeb403c314/common/lib/Form/Class.FormHandler.inc.php#L1457
// https://github.com/Star2Billing/a2billing/blob/8dd474c6077544dcc757159a50149bbeb403c314/common/lib/Form/Class.FormBO.php#L884
// TODO(TobaniEG): create associated invoice
// TODO(TobaniEG): differentiate between client and server errors, for error statuses
func CardRecharge(useralias string, amount int, paymentTxRef string, paymentDate time.Time) *Card {
	db := getDB()
	defer db.Close()

	// use transaction
	tx, err := db.Begin()
	if err != nil {
		utils.Log.Println(err)
		return nil
	}
	defer tx.Rollback()

	card := Card{}
	if err := tx.QueryRow(`
SELECT id,useralias,credit,vat
FROM cc_card WHERE useralias = ?`,
		useralias).Scan(&card.ID, &card.Useralias, &card.Credit, &card.Vat); err != nil {
		utils.Log.Println(err)
		return nil
	}

	paymentType := 0
	createPayment, err := tx.Exec(`
INSERT INTO cc_logpayment (date, payment, card_id, description, added_refill, payment_type)
VALUES (?, ?, ?, ?, ?, ?)
		`, paymentDate, amount, card.ID, "Recharge API "+paymentTxRef, 1, paymentType)
	if err != nil {
		utils.Log.Println(err)
		return nil
	}
	paymentId, _ := createPayment.LastInsertId()

	amountWithoutVat := float64(amount) / (1 + (card.Vat / 100))
	createRefill, err := tx.Exec(`
INSERT INTO cc_logrefill (date, credit, card_id, description, refill_type)
VALUES (?, ?, ?, ?, ?)
	`, paymentDate, amountWithoutVat, card.ID, "Recharge API "+paymentTxRef, paymentType /* refill_type == payment_type */)
	if err != nil {
		utils.Log.Println(err)
		return nil
	}
	refillId, _ := createRefill.LastInsertId()

	// update card credit
	card.Credit += amountWithoutVat
	if _, err := tx.Exec("UPDATE cc_card SET credit = ? WHERE id = ?",
		card.Credit, card.ID); err != nil {
		utils.Log.Println(err)
		return nil
	}

	// link refill and payment
	if _, err := tx.Exec("UPDATE cc_logpayment SET id_logrefill = ? WHERE id = ?",
		refillId, paymentId); err != nil {
		utils.Log.Println(err)
		return nil
	}

	if err := tx.Commit(); err != nil {
		utils.Log.Println(err)
		return nil
	}
	return &card
}
