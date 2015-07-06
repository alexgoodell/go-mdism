package main

import (
	"encoding/csv"
	"fmt"
	"github.com/leesper/go_rng" //imported as rng
	"os"
	"reflect"
	"time"
)

type GammaNumbers struct {
	Value float64
}

func main() {
	gammaGen := rng.NewGammaGenerator(time.Now().UnixNano())
	fmt.Println(gammaGen.Gamma(13.7723, 0.12126))
	InvGammaNumbers := make([]GammaNumbers, 10000, 10000)
	alpha := 13.7723 //These are the variables that you variate in the PSA
	beta := 0.12126
	for i := 0; i < 10000; i++ {
		InvGammaNumbers[i].Value = gammaGen.Gamma(alpha, beta)
	}
	toCsv("GammaNumbers.csv", InvGammaNumbers[1], InvGammaNumbers)

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
