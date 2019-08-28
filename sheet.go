package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"golang.org/x/oauth2/google" // to get sheet client
	"google.golang.org/api/sheets/v4"
)

// spreadsheets clientを取得
func getSheetClient() (*sheets.Service, error) {
	// googleAPIへのclientをリクエストから作成
	client, err := getClientWithJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to getClientWithJSON, %v", err)
	}
	// spreadsheets clientを取得
	srv, err := sheets.New(client)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve Sheets Client %v", err)
	}
	return srv, nil
}

func getClientWithJSON() (*http.Client, error) {
	data, err := ioutil.ReadFile(credentialFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read client secret file. path: '%s', %v", credentialFilePath, err)
	}
	conf, err := google.JWTConfigFromJSON(data, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return nil, fmt.Errorf("failed to parse client secret file to config: %v", err)
	}
	return conf.Client(context.Background()), nil
}

func clearSheet(srv *sheets.Service, sid string, sname string) error {
	// clear stockprice rate spreadsheet:
	resp, err := srv.Spreadsheets.Values.Clear(sid, sname, &sheets.ClearValuesRequest{}).Do()
	if err != nil {
		return fmt.Errorf("failed to clear value. %v", err)
	}
	status := resp.ServerResponse.HTTPStatusCode
	if status != 200 {
		return fmt.Errorf("HTTPstatus error. %v", status)
	}
	return nil
}

// sheetのID, sheet名と対象のデータ（[][]interface{}型）を入力値にとり、
// Sheetにデータを記入する関数
func writeSheet(srv *sheets.Service, sid string, sname string, records [][]interface{}) error {
	valueRange := &sheets.ValueRange{
		MajorDimension: "ROWS",
		Values:         records,
	}
	// Write stockprice rate spreadsheet:
	resp, err := srv.Spreadsheets.Values.Append(sid, sname, valueRange).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		return fmt.Errorf("failed to write value. %v", err)
	}
	status := resp.ServerResponse.HTTPStatusCode
	if status != 200 {
		return fmt.Errorf("HTTPstatus error. %v", status)
	}
	return nil
}

// SheetのClearとWriteを行う関数
func clearAndWriteSheet(srv *sheets.Service, sid string, sname string, records [][]interface{}) error {
	if err := clearSheet(srv, sid, sname); err != nil {
		return fmt.Errorf("failed to clearSheet. sheetID: %s, sheetName: %s, %v", sid, sname, err)
	}

	// writeSheetに渡す
	if err := writeSheet(srv, sid, sname, records); err != nil {
		return fmt.Errorf("failed to writeSheet. sheetID: %s, sheetName: %s, error data: [%v], %v", sid, sname, records, err)
	}
	return nil
}
