package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/gen2brain/beeep"
)

type typ_vox struct {
	IsOp bool `json:"isOp"`
	Vox  struct {
		Title         string `json:"title"`
		Category      int    `json:"category"`
		Filename      string `json:"filename"`
		FileExtension string `json:"fileExtension"`
		IsURL         bool   `json:"isURL"`
		IsBucket      bool   `json:"isBucket"`
		Dice          bool   `json:"dice"`
		Blur          bool   `json:"blur"`
		Sticky        bool   `json:"sticky"`
		URL           string `json:"url"`
		CommentsCount int    `json:"commentsCount"`
		Poll          [][]struct {
			Key   string      `json:"Key"`
			Value interface{} `json:"Value"`
		} `json:"poll"`
		Date        time.Time `json:"date"`
		LastUpdate  time.Time `json:"lastUpdate"`
		Pin         string    `json:"pin"`
		Flag        bool      `json:"flag"`
		Username    string    `json:"username"`
		UniqueID    bool      `json:"uniqueId"`
		Description string    `json:"description"`
	} `json:"vox"`
}

var map_commands map[string]string

func main() {
	fmt.Println(` `)
	var tmp_args []string
	if len(os.Args) > 1 {
		tmp_args = os.Args[1:]
		if slices.Contains([]string{"FOLLOW", "UNFOLLOW"}, tmp_args[0]) && len(os.Args) < 3 {
			exit(errors.New(("sin suficientes parámetros para FOLLOW/UNFOLLOW")))
		}
	} else {
		exit(errors.New(("sin modo de inicio! pon como parámetro HELP/LIST/CHECK/CHECK+/FOLLOW/UNFOLLOW/UNFOLLOW+")))
	}
	map_commands = make(map[string]string)
	map_commands["HELP"] = "ver un listado de los comandos posibles en la aplicación."
	map_commands["LIST"] = "ver un listado de los voxs que actualmente están siendo seguidos."
	map_commands["CHECK"] = "comprobar si hubo comentarios nuevos en el listado de voxs seguidos."
	map_commands["CHECK+"] = "comprobar si hubo comentarios nuevos en el listado de voxs seguidos,\n y generar una notificación de escritorio."
	map_commands["FOLLOW"] = "agregar un listado de voxs al listado de voxs seguidos."
	map_commands["UNFOLLOW"] = "remover un listado de voxs del listado de voxs seguidos."
	map_commands["UNFOLLOW+"] = "remover todo del listado de voxs seguidos."
	os.Args[1] = strings.ToUpper(os.Args[1])
	fmt.Println(`Hola, elegiste el modo ` + os.Args[1] + `, que sirve para ` + map_commands[os.Args[1]] + ``)
	switch os.Args[1] {
	case "HELP":
		help()
	case "LIST":
		list(load())
	case "CHECK":
		check(load(), false)
	case "CHECK+":
		check(load(), true)
	case "FOLLOW":
		follow(load(), tmp_args[1:])
	case "UNFOLLOW":
		unfollow(load(), tmp_args[1:])
	case "UNFOLLOW+":
		unfollow(load(), []string{})
	}
	fmt.Println(` `)
}

func exit(fn_err error) {
	fmt.Println(`Error x_x mira abajo te queda una descripción de lo sucedido. `)
	fmt.Println(fn_err)
	fmt.Println(` `)
	os.Exit(1)
}

func save(fn_voxs []typ_vox) {
	file, err := json.MarshalIndent(fn_voxs, "", " ")
	if err != nil {
		exit(err)
	}
	err = os.WriteFile("./voxs.json", file, 0644)
	if err != nil {
		exit(err)
	}
}

func load() []typ_vox {
	tmp_json, err := os.Open("./voxs.json")
	var tmp_voxs []typ_vox
	if os.IsNotExist(err) {
		file, err := json.MarshalIndent([]typ_vox{}, "", " ")
		if err != nil {
			exit(err)
		}
		err = os.WriteFile("./voxs.json", file, 0644)
		if err != nil {
			exit(err)
		}
	} else if err != nil {
		exit(err)
	} else {
		byteValue, _ := io.ReadAll(tmp_json)
		json.Unmarshal(byteValue, &tmp_voxs)
	}
	return tmp_voxs
}

func already(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func plu(fn_count int) string {
	if fn_count == 1 {
		return "vox. "
	}
	return "voxs. "
}

func follow(fn_voxs []typ_vox, fn_ids []string) {
	fmt.Println(`Ahora se intentará comenzar a seguir un total de ` + fmt.Sprint(len(fn_ids)) + ` ` + plu(len(fn_ids)) + ` `)
	tmp_count, tmp_ids := 0, []string{}
	for i := 0; i < len(fn_voxs); i++ {
		tmp_ids = append(tmp_ids, fn_voxs[i].Vox.Filename)
	}
	for i := 0; i < len(fn_ids); i++ {
		if !already(tmp_ids, fn_ids[i]) {
			tmp_vox := checkOne(fn_ids[i])
			if tmp_vox.Vox.Filename != "" {
				fn_voxs = append(fn_voxs, tmp_vox)
				tmp_count++
			}
		} else {
			fmt.Println(`El identificador del vox ` + fn_ids[i] + ` ya estaba en el listado de seguidos. `)
		}
	}
	if tmp_count > 0 {
		save(fn_voxs)
	}
	fmt.Println(`Se pudo agregar al listado de voxs seguidos un total de ` + fmt.Sprint(tmp_count) + ` ` + plu(tmp_count) + ` `)
}

func unfollow(fn_voxs []typ_vox, fn_ids []string) {
	if len(fn_ids) == 0 {
		for i := 0; i < len(fn_voxs); i++ {
			fn_ids = append(fn_ids, fn_voxs[i].Vox.Filename)
		}
	}
	fmt.Println(`Ahora se intentará dejar de seguir un total de ` + fmt.Sprint(len(fn_ids)) + ` ` + plu(len(fn_ids)) + ` `)
	tmp_voxsNew := []typ_vox{}
	for i := 0; i < len(fn_voxs); i++ {
		if !slices.Contains(fn_ids, fn_voxs[i].Vox.Filename) {
			tmp_voxsNew = append(tmp_voxsNew, fn_voxs[i])
		} else {
			fmt.Println(`El identificador del vox ` + fn_ids[i] + ` no estaba en el listado de seguidos. `)
		}
	}
	if len(fn_voxs) != len(tmp_voxsNew) {
		save(tmp_voxsNew)
	}
	fmt.Println(`Se pudo remover del listado de voxs seguidos un total de ` + fmt.Sprint(len(fn_voxs)-len(tmp_voxsNew)) + ` ` + plu(len(fn_voxs)-len(tmp_voxsNew)) + ` `)
}

func help() {
	fmt.Println(`Ahora verás el listado de los comandos posibles con una explicación de su función. `)
	for k, v := range map_commands {
		fmt.Println("-", k, "para", v)
	}
}

func list(fn_voxs []typ_vox) {
	fmt.Println(`Ahora verás el listado de los voxs que están siendo seguidos actualmente. `)
	for i := 0; i < len(fn_voxs); i++ {
		fmt.Println("-", fn_voxs[i].Vox.Title, "/", fn_voxs[i].Vox.Filename, "/", fn_voxs[i].Vox.CommentsCount)
	}
	fmt.Println(`Se mostró el título / identificador / comentarios de un total de ` + fmt.Sprint(len(fn_voxs)) + ` ` + plu(len(fn_voxs)) + ` `)
}

func check(fn_voxs []typ_vox, fn_notify bool) {
	fmt.Println(`Ahora se comprobarán los comentarios de un total de ` + fmt.Sprint(len(fn_voxs)) + ` ` + plu(len(fn_voxs)) + ` `)
	var tmp_alerts []typ_vox
	var tmp_diff []int
	var tmp_strings string
	for i := 0; i < len(fn_voxs); i++ {
		tmp_voxNow := checkOne(fn_voxs[i].Vox.Filename)
		if fn_voxs[i].Vox.Filename != "" && tmp_voxNow.Vox.LastUpdate.After(fn_voxs[i].Vox.LastUpdate) {
			tmp_alerts = append(tmp_alerts, tmp_voxNow)
			tmp_diff = append(tmp_diff, tmp_voxNow.Vox.CommentsCount-fn_voxs[i].Vox.CommentsCount)
			fn_voxs[i] = tmp_voxNow
		}
	}
	if len(tmp_alerts) > 0 {
		save(fn_voxs)
		for i := 0; i < len(tmp_alerts); i++ {
			tmp_string := "(!!!) " + fmt.Sprint(tmp_diff[i]) + " △ - " + fn_voxs[i].Vox.Title
			tmp_strings += tmp_string + "\n"
			fmt.Printf("%+v\n", tmp_string)
		}
		if fn_notify {
			err := beeep.Notify("Hay actividad", tmp_strings, "assets/information.png")
			if err != nil {
				exit(err)
			}
		}
	}
	fmt.Println(`Se encontraron comentarios nuevos en un total de ` + fmt.Sprint(len(tmp_alerts)) + ` ` + plu(len(tmp_alerts)) + ` `)
}

func checkOne(fn_id string) typ_vox {
	tmp_req, err := http.NewRequest("POST", "https://api.devox.re/voxes/getVox/"+fn_id, nil)
	if err != nil {
		exit(err)
	}
	tmp_r, err := http.DefaultClient.Do(tmp_req)
	if err != nil {
		exit(err)
	}
	var tmp_res typ_vox
	json.NewDecoder(tmp_r.Body).Decode(&tmp_res)
	if tmp_res.Vox.Filename == "" {
		fmt.Println(`El identificador del vox ` + fn_id + ` no pudo ser comprobado. `)
	}
	return tmp_res
}
