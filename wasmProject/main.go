package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var vms map[string]VM = map[string]VM{}

func main() {
	wasm, err := os.ReadFile("./policy.wasm")
	if err != nil {
		log.Panic(err)
	}
	wasm2, err := os.ReadFile("./policy2.wasm")
	if err != nil {
		log.Panic(err)
	}
	data, err := os.ReadFile("./data.json")
	ds := ""
	for _, ch := range data {
		ds += string(ch)
	}
	ds2 := ""
	for _, ch := range ds {
		if string(ch) != "\n" && string(ch) != " " && string(ch) != "\t" {
			ds2 += string(ch)
		}
	}
	if err != nil {
		log.Panic(err)
	}
	opts := vmOpts{wasm, []byte(ds2), 15, &Pool{}}
	opts2 := vmOpts{wasm2, []byte{}, 15, &Pool{}}
	vms["test"] = newVM(opts)
	vms["test2"] = newVM(opts2)
	router := gin.Default()
	router.GET("/data/:vm", getData)
	router.POST("/data/:vm", postData)
	router.GET("/policyWasm/:vm", getPolicyWasm)
	router.POST("/policyWasm/:vm", postPolicyWasm)
	router.POST("/eval/:vm", eval)
	router.Run("192.168.0.113:8080")
}
func getData(c *gin.Context) {
	vm := vms[c.Param("vm")]
	if vm.module.name == "" {
		c.String(http.StatusNotFound, fmt.Sprintf("invalid vm:%s", c.Param("vm")))
		return
	}
	data := vm.getData()
	var jsVal any
	json.Unmarshal([]byte(data), &jsVal)
	c.IndentedJSON(http.StatusOK, jsVal)
}
func postData(c *gin.Context) {
	vm := vms[c.Param("vm")]
	if vm.module.name == "" {
		c.String(http.StatusNotFound, fmt.Sprintf("invalid vm:%s", c.Param("vm")))
		return
	}
	var jsVal any
	c.BindJSON(&jsVal)
	input, _ := json.Marshal(&jsVal)
	out := ""
	for _, ch := range input {
		out += string(ch)
	}
	vm.SetData([]byte(out))
	vm.LoadData()
	vms[c.Param("vm")] = vm
}
func getPolicyWasm(c *gin.Context) {
	vm := vms[c.Param("vm")]
	if vm.module.name == "" {
		c.String(http.StatusNotFound, fmt.Sprintf("invalid vm:%s", c.Param("vm")))
		return
	}
	data := vm.getPolicy()
	c.IndentedJSON(http.StatusOK, data)
}
func postPolicyWasm(c *gin.Context) {
	vm := vms[c.Param("vm")]
	if vm.module.name == "" {
		vm = wasm.newVM(wasm.vmOpts{memoryMin: 15})
	}
	data := c.GetString("policy")
	vm.SetPolicy([]byte(data))
	vm.LoadPolicy()
	vms[c.Param("vm")] = vm
}
func eval(c *gin.Context) {
	log.Println("evaluating")
	vm := vms[c.Param("vm")]
	if vm.module.name == "" {
		c.String(http.StatusNotFound, fmt.Sprintf("invalid vm:%s", c.Param("vm")))
		return
	}
	var jsVal any
	c.BindJSON(&jsVal)
	input, _ := json.Marshal(&jsVal)
	out := ""
	for _, ch := range input {
		out += string(ch)
	}
	json.Unmarshal([]byte(vm.Eval([]byte(out))), &jsVal)
	c.IndentedJSON(http.StatusOK, jsVal)
}
