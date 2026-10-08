package catalog

import (
  _ "embed"
  "encoding/json"
  "fmt"
)

//go:embed models.json
var raw []byte

type Model struct {
  ID string `json:"id"`
  Family string `json:"family"`
  ParamsB float64 `json:"params_b"`
  ActiveB float64 `json:"active_b"`
  Context int `json:"context"`
  Quality float64 `json:"quality"`
  Coding float64 `json:"coding"`
  Reasoning float64 `json:"reasoning"`
  Formats []string `json:"formats"`
  Quants []string `json:"quants"`
}
func All()([]Model,error){var xs []Model;if err:=json.Unmarshal(raw,&xs);err!=nil{return nil,fmt.Errorf("decode model catalog: %w",err)};return xs,nil}
func Find(id string)(Model,bool){xs,_:=All();for _,m:=range xs{if m.ID==id{return m,true}};return Model{},false}
