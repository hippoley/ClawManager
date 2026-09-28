package fit

import (
  "fmt"
  "math"
  "sort"
  "strings"
  "github.com/hippoley/ClawManager/tools/edgefit/internal/catalog"
  "github.com/hippoley/ClawManager/tools/edgefit/internal/hardware"
)
type Result struct {
  ModelID string `json:"model_id"`
  Quant string `json:"quant"`
  Runtime string `json:"runtime"`
  RunPath string `json:"run_path"`
  Fit string `json:"fit"`
  WeightGB float64 `json:"weight_gb"`
  KVGB float64 `json:"kv_gb"`
  RequiredGB float64 `json:"required_gb"`
  PoolGB float64 `json:"pool_gb"`
  HeadroomGB float64 `json:"headroom_gb"`
  EstimatedTPS float64 `json:"estimated_tps"`
  TaskScore float64 `json:"task_score"`
  OverallScore float64 `json:"overall_score"`
  Command string `json:"command"`
  Reason string `json:"reason"`
}
func bytesPerWeight(q string)float64{switch strings.ToUpper(q){case"Q2_K":return .34;case"Q3_K_M":return .46;case"Q4_K_M","Q4_0","AWQ-4BIT","GPTQ-INT4":return .58;case"Q5_K_M":return .72;case"Q6_K":return .84;case"Q8_0","AWQ-8BIT":return 1.05;case"F16","BF16":return 2;default:return .58}}
func memory(m catalog.Model,q string,ctx int)(float64,float64,float64){if ctx<=0{ctx=min(m.Context,8192)};w:=m.ParamsB*bytesPerWeight(q);kv:=math.Max(.18,float64(ctx)/8192*.22*math.Sqrt(math.Max(m.ParamsB,1)));return w,kv,w*1.10+kv+.55}
func Recommend(spec hardware.Specs,models []catalog.Model,task string,ctx int)[]Result{out:=make([]Result,0,len(models));for _,m:=range models{q:="Q4_K_M";if len(m.Quants)>0{q=m.Quants[0]};out=append(out,evaluate(spec,m,q,task,ctx))};sort.Slice(out,func(i,j int)bool{return out[i].OverallScore>out[j].OverallScore});return out}
func Plan(spec hardware.Specs,m catalog.Model,q,task string,ctx int)Result{if q==""{q="Q4_K_M"};return evaluate(spec,m,q,task,ctx)}
func evaluate(spec hardware.Specs,m catalog.Model,q,task string,ctx int)Result{
  w,kv,req:=memory(m,q,ctx);taskScore:=m.Quality;if strings.EqualFold(task,"coding"){taskScore=m.Coding};if strings.EqualFold(task,"reasoning"){taskScore=m.Reasoning}
  gpuPool:=0.0;backend:="";for _,g:=range spec.GPUs{p:=g.FreeVRAMGB;if p<=0{p=g.VRAMGB};gpuPool+=p;if backend==""{backend=g.Backend}}
  ramPool:=spec.AvailableRAMGB;pool:=gpuPool;path:="gpu";runtimeName:="llama.cpp"
  if backend=="metal"{runtimeName="mlx/llama.cpp"};if backend=="cuda"&&contains(m.Formats,"awq"){runtimeName="vLLM/llama.cpp"}
  if gpuPool<=0||req>gpuPool{if gpuPool>0&&req<=gpuPool+ramPool*.80{path="cpu-offload";pool=gpuPool+ramPool*.80}else{path="cpu-only";pool=ramPool;runtimeName="llama.cpp"}}
  ratio:=0.0;if pool>0{ratio=req/pool};level:="too-tight";if ratio>0&&ratio<=.72&&path=="gpu"{level="perfect"}else if ratio>0&&ratio<=.88{level="good"}else if ratio>0&&ratio<=1{level="marginal"}
  head:=pool-req;tps:=estimateTPS(m,q,path,backend,spec.CPUCores);fitScore:=math.Max(0,100-math.Abs(ratio-.72)*120);speedScore:=math.Min(100,tps*2.2);overall:=taskScore*.52+fitScore*.25+speedScore*.23;if level=="too-tight"{overall*=.18}
  cmd:=command(runtimeName,m.ID,q);reason:=fmt.Sprintf("%s via %s; requires %.1f GB, pool %.1f GB, headroom %.1f GB",level,path,req,pool,head)
  return Result{m.ID,q,runtimeName,path,level,round(w),round(kv),round(req),round(pool),round(head),round(tps),round(taskScore),round(overall),cmd,reason}
}
func estimateTPS(m catalog.Model,q,path,backend string,cores int)float64{active:=m.ActiveB;if active<=0{active=m.ParamsB};base:=18/math.Sqrt(math.Max(active,1));switch backend{case"cuda":base*=5.2;case"metal":base*=3.7;default:base*=math.Max(.8,float64(cores)/8)};switch path{case"cpu-offload":base*=.46;case"cpu-only":base*=.24};switch strings.ToUpper(q){case"Q4_K_M","Q4_0":base*=1.18;case"Q5_K_M":base*=1.08;case"Q8_0":base*=.88;case"F16","BF16":base*=.62};return math.Max(.2,base)}
func command(runtimeName,id,q string)string{if strings.HasPrefix(runtimeName,"vLLM"){return fmt.Sprintf("vllm serve %s --dtype auto",id)};if strings.HasPrefix(runtimeName,"mlx"){return fmt.Sprintf("mlx_lm.generate --model %s --prompt \"Hello\"",id)};return fmt.Sprintf("ollama run %s # preferred quant: %s",id,q)}
func contains(xs []string,s string)bool{for _,x:=range xs{if x==s{return true}};return false}
func min(a,b int)int{if a<b{return a};return b}
func round(v float64)float64{return math.Round(v*10)/10}
