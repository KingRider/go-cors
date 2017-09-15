package main

import (
	"bytes"
	"fmt"
	"net/http"
	_ "net/url"
	"os"
	"strings"

	// mapa
	"encoding/xml"
	"io/ioutil"
)

func main() {
	//runtime.GOMAXPROCS(2)

	fmt.Println("Em rodando...\npor Sandro Alvares 2016")

	//os.Setenv("NLS_LANG", "American_America.UTF8")
	os.Setenv("NLS_LANG", "BRAZILIAN PORTUGUESE_BRAZIL.UTF8")

	// handler := cors.Default().Handler(mux)

	http.HandleFunc("/cors", SA_cors)
	http.HandleFunc("/mapa", SA_mapa)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println(err)
		//panic(err)
		os.Exit(1)
	}
}

func SA_cors(rw http.ResponseWriter, req *http.Request) {

	var sa_url = req.FormValue("url")
	if strings.LastIndex(sa_url, ".xml") > 0 || strings.LastIndex(sa_url, "/xml?") > 0 {
		rw.Header().Set("Content-Type", "application/xml; charset=UTF-8")
	} else if strings.LastIndex(sa_url, ".json") > 0 {
		rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
	} else {
		rw.Header().Set("Content-Type", "text/html; charset=UTF-8")
	}

	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Add("Access-Control-Allow-Methods", "GET, PUT, POST, OPTIONS, PATCH, DELETE")
	rw.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	rw.Header().Add("Access-Control-Allow-Credentials", "true")
	rw.Header().Add("Cache-Control", "no-cache")

	var err error
	var jsonData = []byte(req.FormValue("data")) // POST & PUT
	//fmt.Println(string(jsonData)) // SANDRO JSON

	var sa_method = strings.ToUpper(req.FormValue("method"))
	var sa_usuario = req.FormValue("user")
	var sa_senha = req.FormValue("password")
	var sa_mime string = "application/json"

	if sa_url == "" {
		rw.Write([]byte(`SA> Favor digite endereço [url]`))
		/* Leitura de metodo [method] (GET 'padrão', POST e PUT) / Utilizar gravação para [data], [usuario] e [token]`)) */
		return
	}

	if sa_method == "" {
		sa_method = "GET"
	}

	//parametro := strings.LastIndex(sa_url, "?")
	//sa_url[:parametro]
	//sa_url[parametro+1:]
	//sa_url = strings.Replace(sa_url, "%20", "+", -1)
	sa_url = strings.Replace(sa_url, " ", "%20", -1)

	reqJsonQuery, err := http.NewRequest(sa_method, sa_url, bytes.NewBuffer(jsonData))
	if err != nil {
		//http.Error(rw, err.Error(), http.StatusInternalServerError)
		//rw.Write([]byte(`SA> Error: NewRequest<br><br>` + err.Error()))
		fmt.Println("SA> Error NewRequest:", err)
		return
	}
	if sa_usuario != "" && sa_senha != "" {
		reqJsonQuery.SetBasicAuth(sa_usuario, sa_senha)
	}

	if strings.LastIndex(sa_url, ".xml") > 0 || strings.LastIndex(sa_url, "/xml?") > 0 {
		sa_mime = "application/xml"
	}
	//fmt.Println(sa_url)
	reqJsonQuery.Header.Set("Content-Type", sa_mime)
	reqJsonQuery.Close = true

	clientQuery := &http.Client{}
	jsonRespQuery, err := clientQuery.Do(reqJsonQuery)
	if err != nil {
		fmt.Println("SA> Error:", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	b := new(bytes.Buffer)
	b.ReadFrom(jsonRespQuery.Body)
	rw.Write(b.Bytes())

}

// --------------- GOOGLE MAPA

type GeocodeResponse struct {
	Result struct {
		Formatted_address string `xml:"formatted_address"`
		//Address_component []string `xml:"address_component"`
		Geometry struct {
			Location struct {
				Latitude  string `xml:"lat"`
				Longitude string `xml:"lng"`
			} `xml:"location"`
		} `xml:"geometry"`
	} `xml:"result"`
}

func SA_mapa(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Add("Cache-Control", "no-cache")

	var sa_endereco = r.FormValue("endereco")
	sa_endereco = strings.Replace(sa_endereco, "%20", "+", -1)
	sa_endereco = strings.Replace(sa_endereco, " ", "+", -1)

	client := http.Client{}
	sa_xml, err := http.NewRequest("GET", "https://maps.google.com/maps/api/geocode/xml?address="+sa_endereco, nil)
	sa_xml.Header.Set("Content-Type", "application/xml")
	sa_xml.Header.Add("Cache-Control", "no-cache")
	sa_xml.Header.Add("Connection", "close")

	sa_readxml, err := client.Do(sa_xml)
	if err != nil {
		fmt.Println("SA> Error:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if sa_readxml != nil {
		defer sa_readxml.Body.Close()
	}

	corpo, err := ioutil.ReadAll(sa_readxml.Body)
	if err != nil {
		fmt.Println("SA> Error:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var mapaXML GeocodeResponse
	//fmt.Println(string(corpo))
	err = xml.Unmarshal(corpo, &mapaXML)
	//fmt.Println(mapaXML)
	if err != nil {
		fmt.Println("SA> Error:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//json.NewEncoder(w).Encode(mapaXML)
	json := []byte(`{
		"endereco": "` + mapaXML.Result.Formatted_address + `",
		"latitude": "` + mapaXML.Result.Geometry.Location.Latitude + `",
		"longitude": "` + mapaXML.Result.Geometry.Location.Longitude + `"
	}`)
	w.Write(json)

}

// --------------- FIM MAPA
