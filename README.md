# A2Billing API

A2Billing API is a web service.

Routes:
- GET    /ping
- GET    /api/clientbalance
- POST   /api/clientrecharge

## Usage

Make sure you have [golang installed](https://go.dev/doc/install):

```sh
git clone https://github.com/ITSky-Solutions/a2billing-api-go.git
cd a2billing-api-go
make start # builds and starts the server
```
## Environment Variables

| Name        | Description | Required |
| ----------- | ----------- | ----------- |
| API_DB_USER | DB username | :white_check_mark: |
| API_DB_PASS | DB password | :x: |
| API_DB_NAME | DB name | :white_check_mark: |
| API_DB_HOST | DB host | :white_check_mark: |
| API_DB_PORT | DB port | :x: |
| API_PORT    | Port | :x: |
| API_KEY     | Client Auth Key | :white_check_mark: |
