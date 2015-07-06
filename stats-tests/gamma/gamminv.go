// Gamma distribution
// k > 0                shape parameter
// θ (Theta) > 0        scale parameter

package main

import (
	. "code.google.com/p/go-fn/fn"
	"encoding/csv"
	"fmt"
	"math"
	"math/rand"
	"os"
	"reflect"
	"time"
)

type GammaNumbers struct {
	Value float64
}

func main() {
	rand.Seed(time.Now().Unix())
	random := rand.Float64()
	fmt.Println(Gamma_InvCDF_For(13.7723, 0.12126, random))
	InvGammaNumbers := make([]GammaNumbers, 10000, 10000)
	i := 0
	for i < 10000 {
		random = rand.Float64()
		InvGammaNumbers[i].Value = Gamma_InvCDF_For(13.7723, 0.12126, random)
		i = i + 1
	}

	toCsv("GammaNumbers", InvGammaNumbers[1], InvGammaNumbers)

}

// Probability density function
func Gamma_PDF(k float64, θ float64) func(x float64) float64 {
	return func(x float64) float64 {
		if x < 0 {
			return 0
		}
		return math.Pow(x, k-1) * math.Exp(-x/θ) / (Γ(k) * math.Pow(θ, k))
	}
}

// Cumulative distribution function, analytic solution, did not pass some tests!
func Gamma_CDF(k float64, θ float64) func(x float64) float64 {
	return func(x float64) float64 {
		if k < 0 || θ < 0 {
			panic(fmt.Sprintf("k < 0 || θ < 0"))
		}
		if x < 0 {
			return 0
		}
		return Iγ(k, x/θ) / Γ(k)
	}
}

// Value of the probability density function at x
func Gamma_PDF_At(k, θ, x float64) float64 {
	pdf := Gamma_PDF(k, θ)
	return pdf(x)
}

// Value of the cumulative distribution function at x
func Gamma_CDF_At(k, θ, x float64) float64 {
	cdf := Gamma_CDF(k, θ)
	return cdf(x)
}

// Inverse CDF (Quantile) function
func Gamma_InvCDF(k float64, θ float64) func(x float64) float64 {
	return func(x float64) float64 {
		var eps, y_new, h float64
		eps = 1e-4
		y := k * θ
		y_old := y
	L:
		for i := 0; i < 100; i++ {
			h = (Gamma_CDF_At(k, θ, y_old) - x) / Gamma_PDF_At(k, θ, y_old)
			y_new = y_old - h
			if y_new <= eps {
				y_new = y_old / 10
				h = y_old - y_new
			}
			if math.Abs(h) < eps {
				break L
			}
			y_old = y_new
		}
		return y_new
	}
}

// Value of the inverse CDF for probability p
func Gamma_InvCDF_For(k, θ, p float64) float64 {
	cdf := Gamma_InvCDF(k, θ)
	return cdf(p)
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
