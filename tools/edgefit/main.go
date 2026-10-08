package main

import (
  "encoding/json"
  "flag"
  "fmt"
  "os"
  "strings"
  "github.com/hippoley/ClawManager/tools/edgefit/internal/bench"
  "github.com/hippoley/ClawManager/tools/edgefit/internal/catalog"
  "github.com/hippoley/ClawManager/tools/edgefit/internal/fit"
  "github.com/hippoley/ClawManager/tools/edgefit/internal/hardware"
)
func main(){
  if len(os.Args)<2{usage();os.Exit(2)}
  switch os.Args[1]{case"scan":scan(os.Args[2:]);case"list":list(os.Args[2:]);case"recommend":recommend(os.Args[2:]);case"plan":plan(os.Args[2:]);case"doctor":doctor(os.Args[2:]);case"bench":benchCmd(os.Args[2:]);default:usage();os.Exit(2)}
}
func usage(){fmt.Println("edgefit — local model hardware planner\n\nUsage:\n  edgefit scan [--json]\n  edgefit list [--json]\n  edgefit recommend [--task general|coding|reasoning] [--context 8192] [--top 8] [--json]\n  edgefit plan --model <id> [--quant Q4_K_M] [--task coding] [--context 8192] [--json]\n  edgefit doctor [--json]\n  edgefit bench --model <ollama-model> [--url http://127.0.0.1:11434] [--json]")}
func scan(args []string){fs:=flag.NewFlagSet("scan",flag.ExitOnError);js:=fs.Bool("json",false,"");_ = fs.Parse(args);output(hardware.Detect(),*js)}
func list(args []string){fs:=flag.NewFlagSet("list",flag.ExitOnError);js:=fs.Bool("json",false,"");_ = fs.Parse(args);xs,err:=catalog.All();die(err);if *js{output(xs,true);return};for _,m:=range xs{fmt.Printf("%-34s %5.1fB ctx=%-7d %s\n",m.ID,m.ParamsB,m.Context,strings.Join(m.Quants,","))}}
func recommend(args []string){fs:=flag.NewFlagSet("recommend",flag.ExitOnError);task:=fs.String("task","general","");ctx:=fs.Int("context",8192,"");top:=fs.Int("top",8,"");js:=fs.Bool("json",false,"");_ = fs.Parse(args);xs,err:=catalog.All();die(err);rs:=fit.Recommend(hardware.Detect(),xs,*task,*ctx);if *top<len(rs){rs=rs[:*top]};if *js{output(rs,true);return};fmt.Printf("%-34s %-9s %-11s %-13s %7s %7s %7s\n","MODEL","FIT","PATH","RUNTIME","GB","TPS","SCORE");for _,r:=range rs{fmt.Printf("%-34s %-9s %-11s %-13s %7.1f %7.1f %7.1f\n",r.ModelID,r.Fit,r.RunPath,r.Runtime,r.RequiredGB,r.EstimatedTPS,r.OverallScore)}}
func plan(args []string){fs:=flag.NewFlagSet("plan",flag.ExitOnError);id:=fs.String("model","","");q:=fs.String("quant","Q4_K_M","");task:=fs.String("task","general","");ctx:=fs.Int("context",8192,"");js:=fs.Bool("json",false,"");_ = fs.Parse(args);if *id==""{die(fmt.Errorf("--model is required"))};m,ok:=catalog.Find(*id);if !ok{die(fmt.Errorf("unknown model %q",*id))};r:=fit.Plan(hardware.Detect(),m,*q,*task,*ctx);if *js{output(r,true);return};fmt.Printf("Model:      %s\nFit:        %s\nRun path:   %s\nRuntime:    %s\nMemory:     %.1f GB (weights %.1f + KV %.1f + overhead)\nPool:       %.1f GB\nHeadroom:   %.1f GB\nEst. TPS:   %.1f\nTask score: %.1f\nWhy:        %s\nCommand:    %s\n",r.ModelID,r.Fit,r.RunPath,r.Runtime,r.RequiredGB,r.WeightGB,r.KVGB,r.PoolGB,r.HeadroomGB,r.EstimatedTPS,r.TaskScore,r.Reason,r.Command)}
func doctor(args []string){fs:=flag.NewFlagSet("doctor",flag.ExitOnError);js:=fs.Bool("json",false,"");_ = fs.Parse(args);names:=[]string{"ollama","llama-cli","nvidia-smi","rocminfo","docker","python3"};m:=map[string]bool{};for _,n:=range names{m[n]=hardware.CommandExists(n)};if *js{output(m,true);return};for _,n:=range names{s:="missing";if m[n]{s="ok"};fmt.Printf("%-12s %s\n",n,s)}}
func benchCmd(args []string){fs:=flag.NewFlagSet("bench",flag.ExitOnError);model:=fs.String("model","","");url:=fs.String("url","http://127.0.0.1:11434","");js:=fs.Bool("json",false,"");_ = fs.Parse(args);if *model==""{die(fmt.Errorf("--model is required"))};r,err:=bench.Ollama(*url,*model);die(err);if *js{output(r,true);return};fmt.Printf("%s: %.1f tok/s, TTFT %.1f ms, total %.1f ms (%d tokens)\n",r.Model,r.TPS,r.TTFTMS,r.TotalMS,r.OutputTokens)}
func output(v any,js bool){if js{b,_:=json.MarshalIndent(v,"","  ");fmt.Println(string(b));return};fmt.Printf("%+v\n",v)}
func die(err error){if err!=nil{fmt.Fprintln(os.Stderr,"edgefit:",err);os.Exit(1)}}
