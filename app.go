package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)


type App struct{
	Router *mux.Router
	DB *sql.DB
}

func (app *App)Initialise(DBUser string,DBPassword string,DBName string)error{
	connectionString :=fmt.Sprintf("%v:%v@tcp(127.0.0.1:3306)/%v",DBUser,DBPassword,DBName)
	var err error
	app.DB,err = sql.Open("mysql",connectionString)
	if err!=nil{
		return err
	}
	app.Router =mux.NewRouter().StrictSlash(true)
	app.handleRoutes()
	return nil
}
func (app *App)Run(address string){
	log.Fatal(http.ListenAndServe(address,app.Router))
}
func sendResponse(w http.ResponseWriter,statusCode int,payload interface{}){
	response,_ :=json.Marshal(payload)
	w.Header().Set("Content-type","application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}
func sendError(w http.ResponseWriter,statusCode int,err string){
	error_message :=map[string]string{"string": err}
	sendResponse(w,statusCode,error_message)
}
func (app *App)getProducts(w http.ResponseWriter,r *http.Request){
	products,err := getProducts(app.DB)
	if err != nil{
		sendError(w,http.StatusInternalServerError,err.Error())
		return
	}
	sendResponse(w,http.StatusCreated,products)

}

func (app *App)getProduct(w http.ResponseWriter,r *http.Request){
	vars :=mux.Vars(r)
	key,err :=strconv.Atoi(vars["id"])
	if err!=nil{
		sendError(w,http.StatusBadRequest,"Invalid product id")
		return
	}
	p :=product{ID:key}
	err = p.getProduct(app.DB)
	if err!=nil{
		switch err{
		case sql.ErrNoRows:
			sendError(w,http.StatusNotFound,"Product not found")
		default:
			sendError(w,http.StatusInternalServerError,err.Error())

		}
		return
	}
	sendResponse(w,http.StatusOK,p)



}

func(app *App)addProduct(w http.ResponseWriter,r *http.Request){
    var p product

	err := json.NewDecoder(r.Body).Decode(&p)
	if err!=nil{
		sendError(w,http.StatusBadRequest,"Invalid request payload")
		return
	}
	err = p.addProduct(app.DB)
	if err!=nil{
		sendError(w,http.StatusInternalServerError,err.Error())
		return
	}
	sendResponse(w,http.StatusCreated,p)

}
func(app *App)replaceProduct(w http.ResponseWriter,r *http.Request){
	vars :=mux.Vars(r)
	key,err := strconv.Atoi(vars["id"])
	if err!=nil{
		sendError(w,http.StatusBadRequest,"ID Doesn't exist")
		return
	}
	var p product

	err = json.NewDecoder(r.Body).Decode(&p)
	if err!=nil{
		sendError(w,http.StatusBadRequest,"Invalid request payload")
		return
	}
	p.ID = key
	err = p.replaceProduct(app.DB)
	if err!=nil{
		sendResponse(w,http.StatusInternalServerError,err.Error())
		return
	}
	sendResponse(w,http.StatusOK,p)
}
func (app *App)deleteProduct(w http.ResponseWriter,r *http.Request){
	vars :=mux.Vars(r)
	key,err := strconv.Atoi(vars["id"])
	if err!=nil{
		sendError(w,http.StatusBadRequest,"ID Doesn't exist")
		return
	}
	p := product{ID:key}
	err = p.deleteProduct(app.DB)
	if err !=nil{
		sendError(w,http.StatusInternalServerError,err.Error())
		return
	}
	sendResponse(w,http.StatusOK,map[string]string{"result":"successfully deleted"})

}
func (app *App)handleRoutes(){
	app.Router.HandleFunc("/products",app.getProducts).Methods("GET")
	app.Router.HandleFunc("/products/{id}",app.getProduct).Methods("GET")
	app.Router.HandleFunc("/addproduct",app.addProduct).Methods("POST")
	app.Router.HandleFunc("/replaceproduct/{id}",app.replaceProduct).Methods("PUT")
	app.Router.HandleFunc("/deleteproduct/{id}",app.deleteProduct).Methods("DELETE")
}