package main

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

//Config - DBType field is equal to Config.toml [header]
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBname   string
	Coll     string
	DBtype   string
}

func (c *Config) read() {
	conf := viper.New()
	conf.SetConfigName("Config")
	conf.AddConfigPath("./")
	err := conf.ReadInConfig()
	if err != nil {
		log.Fatalf("Cannot find Config.toml or file incorrect - %s\n", err)
	}
	s := reflect.ValueOf(c).Elem()
	for i := 0; i < s.NumField(); i++ {
		v := s.Field(i)
		if v.IsValid() {
			//assign to struct field proper Config.toml field
			if val, ok := conf.Get(fmt.Sprintf("%s.%s", c.DBtype, strings.ToLower(s.Type().Field(i).Name))).(string); ok {
				v.SetString(val)
			}
		}

	}
}

//DB struct for CRUD operations
type DB struct {
	ConPool Connection
	//DBType  string
	Config Config
}

//Connection for different dbs
type Connection interface {
	connect(Config) error
	insert(interface{}) error
	close() error
}

//Connect to Database
func (db *DB) Connect(con Connection) error {
	db.ConPool = con
	return db.ConPool.connect(db.Config)
}

//Connect to Database
func (db *DB) Close() {
	db.ConPool.close()
	return
}

//Insert to DB
func (db *DB) Insert(data interface{}) error {
	return db.ConPool.insert(data)
}

//NewDB is a DB constructor with Config.toml configuration
func NewDB(dbType string) *DB {
	conf := Config{DBtype: dbType}
	conf.read()
	return &DB{Config: conf}
}

//PSQLconnector - custom realization of db connector
type PSQLconnector struct {
	db *sql.DB
}

func (psql *PSQLconnector) connect(conf Config) error {
	log.Println("Postgres is connecting... ")
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", conf.Host, conf.Port, conf.User, conf.Password, conf.DBname))
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		log.Printf("Error!%v", err)
		log.Println("Retry Postgres connection in 15 seconds... ")
		time.Sleep(time.Duration(15) * time.Second)
		return psql.connect(conf)
	}
	psql.db = db
	log.Println("Postgres is connected ")
	return nil
}

func (psql *PSQLconnector) close() error {
	err := psql.db.Close()
	if err != nil {
		return err
	}
	log.Println("Closing connection to Postgres... ")
	return nil
}

//Model is Data model struct
type DataModel struct {
	FirstName  string
	SecondName string
}

func (psql *PSQLconnector) insert(data interface{}) error {
	switch data.(type) {
	case DataModel:
		m := data.(DataModel)
		_, err := psql.db.Exec("INSERT INTO example(FIRSTNAME,SECONDNAME) VALUES($1,$2)", m.FirstName, m.SecondName)
		if err != nil {
			return err
		}
	}
	return nil
}
