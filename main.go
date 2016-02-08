package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

const (
	CSVFormat   = "\"%s\",\"%s\",\"%s\",\"%s\"\n"
	QueryMake   = "http://veiculos.fipe.org.br/api/veiculos/ConsultarMarcas"
	QueryModels = "http://veiculos.fipe.org.br/api/veiculos/ConsultarModelos"
	Language    = "portuguese"
)

type QueryResult struct {
	Label string
	Value string
}

type QueryResultInt struct {
	Label string
	Value int
}

type QueryModelYear struct {
	Models []QueryResultInt `json:"Modelos"`
	Years  []QueryResult    `json:"Anos"`
}

type VehicleModel struct {
	Type     string `json:type`
	Make     string `json:make`
	Model    string `json:model`
	Language string `json:language`
}

func main() {
	types := []QueryResult{
		QueryResult{"car", "1"},
		QueryResult{"motorcycle", "2"},
		QueryResult{"heavy", "3"},
	}
	result := make([]VehicleModel, 0)
	counter := 0
	fmt.Print("Found: ")

	for _, vType := range types {
		resp, err := http.PostForm(QueryMake,
			url.Values{
				"codigoTabelaReferencia": {"188"},
				"codigoTipoVeiculo":      {vType.Value},
			})
		if err != nil {
			fmt.Printf("Error querying make: %v\n", err)
			continue
		}

		makes := make([]QueryResult, 0)
		if err := json.NewDecoder(resp.Body).Decode(&makes); err != nil {
			resp.Body.Close()
			fmt.Printf("Error decoding make response: %v\n", err)
			continue
		}
		resp.Body.Close()

		for _, vMake := range makes {
			resp, err = http.PostForm(QueryModels,
				url.Values{
					"codigoTabelaReferencia": {"188"},
					"codigoTipoVeiculo":      {vType.Value},
					"codigoMarca":            {vMake.Value},
				})
			if err != nil {
				fmt.Printf("Error querying model: %v\n", err)
				continue
			}
			models := QueryModelYear{
				make([]QueryResultInt, 0),
				make([]QueryResult, 0),
			}
			if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
				resp.Body.Close()
				fmt.Printf("Error decoding model response: %v\n", err)
				continue
			}
			resp.Body.Close()
			counter += len(models.Models)
			fmt.Print(counter, " ")

			for _, vModel := range models.Models {
				result = append(result, VehicleModel{
					Type:     vType.Label,
					Make:     vMake.Label,
					Model:    vModel.Label,
					Language: Language,
				})
			}
		}
	}
	fmt.Println()

	f, err := os.OpenFile("result.json",
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		fmt.Printf("Could not open/create JSON file: %v\n", err)
		return
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(result); err != nil {
		fmt.Printf("Could not encode JSON object: %v\n", err)
		return
	}

	f, err = os.OpenFile("result.csv",
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		fmt.Printf("Could not open/create CSV file: %v\n", err)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, CSVFormat,
		"Type", "Make", "Model", "Language")
	for _, line := range result {
		fmt.Fprintf(f, CSVFormat,
			line.Type, line.Make, line.Model, line.Language)
	}
}
