package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"time"
)

type NormalNumbers struct {
	Value float64
}

func main() {
	rand.Seed(time.Now().UnixNano())
	invNormalNumbers := make([]NormalNumbers, 10000, 10000)
	mean := 8000 //These are the two input parameters for the PSA
	standardDev := 500
	for i := 0; i < 10000; i++ {
		invNormalNumbers[i].Value = mean + standardDev*rand.NormFloat64()
	}

	toCsv("NormalNumbers.csv", invNormalNumbers[1], invNormalNumbers)

}

func toCsv(filename string, record interface{}, records interface{}) error {
	fmt.Println("Beginning export process to ", filename)
	//create or open file
	os.Create(filename)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	// new Csv wriier
	writer := csv.NewWriter(file)
	// use the single record to determine the fields of the struct
	val := reflect.Indirect(reflect.ValueOf(record))
	numberOfFields := val.Type().NumField()
	var fieldNames []string
	for i := 0; i < numberOfFields; i++ {
		fieldNames = append(fieldNames, val.Type().Field(i).Name)
	}
	// print field names of struct
	err = writer.Write(fieldNames)
	// print the values from the array of structs
	val2 := reflect.ValueOf(records)
	for i := 0; i < val2.Len(); i++ {
		var line []string
		for p := 0; p < numberOfFields; p++ {
			//convert interface to string
			line = append(line, fmt.Sprintf("%v", val2.Index(i).Field(p).Interface()))
		}
		err = writer.Write(line)
	}
	if err != nil {
		fmt.Println("error")
		os.Exit(1)
	}
	fmt.Println("Exported to ", filename)
	writer.Flush()
	return err
}
