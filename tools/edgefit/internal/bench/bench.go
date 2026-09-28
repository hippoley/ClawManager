package bench

import (
  "bytes"
  "encoding/json"
  "fmt"
  "net/http"
  "strings"
  "time"
)
type Result struct {
  Provider string `json:"provider"`
  Model string `json:"model"`
  TPS float64 `json:"tokens_per_second"`
  TTFTMS float64 `json:"ttft_ms"`
  TotalMS float64 `json:"total_ms"`
  OutputTokens int `json:"output_tokens"`
}
type ollamaResponse struct {
  EvalCount int64 `json:"eval_count"`
  EvalDuration int64 `json:"eval_duration"`
  PromptEvalDuration int64 `json:"prompt_eval_duration"`
  TotalDuration int64 `json:"total_duration"`
}
func Ollama(baseURL,model string)(Result,error){
  if baseURL==""{baseURL="http://127.0.0.1:11434"};url:=strings.TrimRight(baseURL,"/")+"/api/generate"
  payload,_:=json.Marshal(map[string]any{"model":model,"prompt":"Explain in three bullets why local inference can be useful.","stream":false,"options":map[string]any{"num_predict":96}})
  start:=time.Now();resp,err:=http.Post(url,"application/json",bytes.NewReader(payload));if err!=nil{return Result{},err};defer resp.Body.Close()
  if resp.StatusCode/100!=2{return Result{},fmt.Errorf("ollama returned %s",resp.Status)}
  var r ollamaResponse;if err:=json.NewDecoder(resp.Body).Decode(&r);err!=nil{return Result{},err}
  wall:=time.Since(start);tps:=0.0;if r.EvalCount>0&&r.EvalDuration>0{tps=float64(r.EvalCount)/(float64(r.EvalDuration)/1e9)}
  total:=float64(r.TotalDuration)/1e6;if total==0{total=float64(wall.Milliseconds())}
  return Result{Provider:"ollama",Model:model,TPS:tps,TTFTMS:float64(r.PromptEvalDuration)/1e6,TotalMS:total,OutputTokens:int(r.EvalCount)},nil
}
