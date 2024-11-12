package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

var a App

func TestMain(m *testing.M){
	err := a.Initialise(DBUser,DBPassword,"test")
	if err!=nil{
		log.Fatal("error occured while initialising the database")
	}
	createTable()
	m.Run()
}
func createTable(){
	createTableQuery :=`CREATE TABLE IF NOT EXISTS products(
		id int NOT NULL AUTO_INCREMENT,
		name varchar(255) NOT NULL,
		quantity int,
		price float(10,7),
		PRIMARY KEY (id)
	);`
	_,err:=a.DB.Exec(createTableQuery)
	if err!=nil{
		log.Fatal(err)
	}
}
func clearTable(){
	a.DB.Exec("DELETE from products")
	a.DB.Exec("ALTER table products AUTO_INCREMENT=1")
	log.Println("clearTable")
}
func addProduct(name string,quantity int,price float64){
	query:=fmt.Sprintf("INSERT into products(name,quantity,price) VALUES('%v','%v','%v')",name,quantity,price)
	_,err := a.DB.Exec(query)
	if err !=nil{
		log.Println(err)
	}

}
func TestGetProduct(t *testing.T){
	clearTable()
	addProduct("keyboard",100,500.00)
	request,_:=http.NewRequest("GET","/products/1",nil)
	response:= sendRequest(request)
	checkStatusCode(t,http.StatusOK,response.Code)
}
func checkStatusCode(t *testing.T,expectedStatusCode int,actualStatusCode int){
	if expectedStatusCode!=actualStatusCode{
		t.Errorf("Expected status: %v,Recieved: %v",expectedStatusCode,actualStatusCode)
	}
}

func sendRequest(r *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	a.Router.ServeHTTP(recorder,r)
	return  recorder
}
func TestCreateProduct(t *testing.T){
	clearTable()
	var product = []byte(`{"name": "chair", "quantity": 2, "price": 600}`)
	req,_:= http.NewRequest("POST","/addproduct",bytes.NewBuffer(product))
	req.Header.Set("Content-Type","application/json")
	response := sendRequest(req)
	checkStatusCode(t,http.StatusCreated,response.Code)
	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(),&m)
	if m["name"]!="chair"{
		t.Errorf("Expected name: %v,Got: %v","chair",m["name"])
	}
	if m["quantity"]!= 2.0{
		t.Errorf("Expected quantity: %v,Got: %v",2,m["quantity"])
	}
}
func TestDeleteProduct(t *testing.T){
	clearTable()
	addProduct("mouse",10,200)
	req,_:=http.NewRequest("GET","/products/1",nil)
	response := sendRequest(req)
	checkStatusCode(t,http.StatusOK,response.Code)

	req,_=http.NewRequest("DELETE","/deleteproduct/1",nil)
	response = sendRequest(req)
	checkStatusCode(t,http.StatusOK,response.Code)

	req,_ =http.NewRequest("GET","/products/1",nil)
	response = sendRequest(req)
	checkStatusCode(t,http.StatusNotFound,response.Code)
}
func TestReplaceProduct(t *testing.T){
	clearTable() 
	addProduct("mouse",10,200)

	req,_:=http.NewRequest("GET","/products/1",nil)
	response := sendRequest(req)

	var oldValue map[string]interface{}
	json.Unmarshal(response.Body.Bytes(),&oldValue)

	var product = []byte(`{"name": "mouse", "quantity": 10, "price": 600}`)
	req,_= http.NewRequest("PUT","/replaceproduct/1",bytes.NewBuffer(product))
	req.Header.Set("Content-Type","application/json")
	response = sendRequest(req)

	var newValue map[string]interface{}
	json.Unmarshal(response.Body.Bytes(),&newValue)

	if oldValue["id"] != newValue["id"]{
		t.Errorf("Expected id: %v, New id:%v",newValue["id"],oldValue["id"])
	}

	if oldValue["name"] != newValue["name"]{
		t.Errorf("Expected name: %v, New name:%v",newValue["name"],oldValue["name"])
	}
	if oldValue["price"] == newValue["price"]{
		t.Errorf("Expected price: %v, New price:%v",newValue["price"],oldValue["price"])
	}
	if oldValue["quantity"] != newValue["quantity"]{
		t.Errorf("Expected quantity: %v, New quantity:%v",newValue["quantity"],oldValue["quantity"])
	}
	
}
