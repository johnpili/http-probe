package controllers

import (
	"fmt"
	"log"
	"reflect"

	"github.com/go-zoo/bone"
)

// RequestMapping ...
func (z *Hub) RequestMapping(router *bone.Mux, requestMapping RequestMapping) {
	requestMapping.RequestMapping(router)
}

// BindRequestMapping ...
func (z *Hub) BindRequestMapping(router *bone.Mux) {
	log.Println("Binding RequestMapping for:")
	for _, v := range z.Controllers {
		z.RequestMapping(router, v.(RequestMapping))
		rt := reflect.TypeOf(v)
		log.Println(rt)
	}
	fmt.Println("")

	log.Println("Binded RequestMapping are the following: ")
	for _, v := range router.Routes {
		for _, m := range v {
			log.Println(m.Method, " : ", m.Path)
		}
	}
	fmt.Println("")
}
