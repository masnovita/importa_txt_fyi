package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var NParaules, NErrors int
var Castellano []string
var Ingles []string
var LCas, LEng int
var separador string

var id_unidad int
var id_usuario int
var front string
var back string

func main() {

	// llegeix fitxers txt que tenim al directori
	ruta, errRuta := os.Getwd()
	if errRuta != nil {
		fmt.Println("Error:", errRuta)
		return
	}

	files, errReadDir := ioutil.ReadDir(ruta)
	if errReadDir != nil {
		fmt.Println("Error:", errReadDir)
		return
	}
	var txtFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".txt") {
			txtFiles = append(txtFiles, file.Name())
		}
	}

	if len(txtFiles) == 0 {
		fmt.Println("No s'han trovat txt en el directori.")
		return
	}
	fmt.Println("Txt trovats :")
	for i, file := range txtFiles {
		fmt.Printf("%d. %s\n", i+1, file)
	}

	var seleccion int
	fmt.Print("Selecciona un arxiu: ")
	fmt.Scanln(&seleccion)

	if seleccion < 1 || seleccion > len(txtFiles) {
		fmt.Println("Selección inválida.")
		return
	}

	arxiuSeleccionat := txtFiles[seleccion-1]
	fmt.Printf("Archivo seleccionado: %s\n", arxiuSeleccionat)

	// llegeix variables d'entorn
	user := os.Getenv("MYSQL_USER")
	password := os.Getenv("MYSQL_PASSWORD")
	database := os.Getenv("MYSQL_DATABASE")
	host := os.Getenv("MYSQL_HOST")

	var liniaOk string // aqui guardarem la linea amb les paraules en castellà i angles

	fmt.Print("Id unitat: ")
	_, errUnitat := fmt.Scan(&id_unidad)
	if errUnitat != nil {
		fmt.Println("Error:", errUnitat)
		return
	}
	fmt.Print("Id usuari: ")
	_, errUsuari := fmt.Scan(&id_usuario)
	if errUsuari != nil {
		fmt.Println("Error:", errUsuari)
		return
	}

	file, err := os.Open(arxiuSeleccionat)                         // obrir fitxer d'entrada
	fileDesti, err2 := os.Create(arxiuSeleccionat + "reparat.txt") // Creem  fitxer reparat
	if err != nil || err2 != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Connecta a la bbdd
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, host, database)
	Db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Println("error mysl")
		panic(err.Error() + " Error connexió")
	}
	defer Db.Close()

	scanner := bufio.NewScanner(file)

	Castellano = []string{}
	Ingles = []string{}
	LCas = 0
	LEng = 0

	for scanner.Scan() {
		palabras := strings.Fields(scanner.Text()) // llegeix una linea
		LCas, LEng = treuParaules(palabras)        // separa la linea en paraules

		if LEng == 0 { // si les paraules en angles estan en un altre linea, torna a llegir
			continue
		} else {
			// guarda a fitxer reparat i posa separador entre castella i angles
			separador = " "
			liniaOk = strings.Join(Castellano, "") + separador + strings.Join(Ingles, "")
			_, err = fileDesti.WriteString(liniaOk + "\n")
			if err != nil {
				fmt.Println("error al guardar fitxer reparat")
			}

			// guarda a la BBDD
			front = strings.Join(Castellano, "")
			back = strings.Join(Ingles, "")
			_, err = Db.Exec("INSERT INTO tarjetas (id_unidad,id_usuario, front,back) VALUES (?, ?, ?, ?)", id_unidad, id_usuario, front, back)
			if err != nil {
				fmt.Println("error INSERT")
				panic(err.Error())
			}

			Castellano = []string{}
			Ingles = []string{}
			LCas = 0
			LEng = 0
		}

		NParaules++

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("%d Errors\n", NErrors)
	fmt.Printf("Fitxer generat : bikes_reparat.txt")
	fmt.Printf(" %d Paraules insertades Mysql \n", NParaules)

	// guarda log a fitxer reparat
	log := strconv.Itoa(NParaules) + " Paraules insertades Mysql. Unitat : " + strconv.Itoa(id_unidad) + "  Usuari : " + strconv.Itoa(id_usuario)
	_, err = fileDesti.WriteString("\n" + log + "\n")
	if err != nil {
		fmt.Println("error al guardar fitxer reparat")
	}

	file.Close()
	fileDesti.Close()

}

func treuParaules(palabras []string) (int, int) {
	var estado string
	if len(palabras) > 0 && palabras[0] == strings.ToLower(palabras[0]) {
		estado = "Castellano"
	} else {
		estado = "Ingles"
	}

	for _, palabra := range palabras {
		if estado == "Castellano" && (strings.ToLower(palabra) == palabra || isNumero(palabra)) {
			Castellano = append(Castellano, palabra+" ")
		} else if estado == "Castellano" {
			estado = "Ingles"
			Ingles = append(Ingles, strings.ToUpper(palabra)+" ")
		} else if estado == "Ingles" && (strings.ToUpper(palabra) == palabra || isNumero(palabra)) {
			Ingles = append(Ingles, palabra+" ")
		} else if estado == "Ingles" {
			Castellano = append(Castellano, palabra+" ")
			estado = "Castellano"
		}
	}
	return len(Castellano), len(Ingles)
}

func isNumero(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
