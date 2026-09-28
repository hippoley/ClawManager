package hardware

import (
  "bufio"
  "bytes"
  "fmt"
  "os"
  "os/exec"
  "runtime"
  "strconv"
  "strings"
)

type GPU struct {
  Name string `json:"name"`
  VRAMGB float64 `json:"vram_gb"`
  FreeVRAMGB float64 `json:"free_vram_gb"`
  Backend string `json:"backend"`
  Unified bool `json:"unified_memory"`
}
type Specs struct {
  OS string `json:"os"`
  Arch string `json:"arch"`
  CPU string `json:"cpu"`
  CPUCores int `json:"cpu_cores"`
  TotalRAMGB float64 `json:"total_ram_gb"`
  AvailableRAMGB float64 `json:"available_ram_gb"`
  GPUs []GPU `json:"gpus"`
}

func Detect() Specs {
  s:=Specs{OS:runtime.GOOS,Arch:runtime.GOARCH,CPUCores:runtime.NumCPU()}
  s.CPU=detectCPU(); s.TotalRAMGB,s.AvailableRAMGB=detectRAM(); s.GPUs=detectGPUs(s.TotalRAMGB)
  return s
}
func detectCPU() string {
  if runtime.GOOS=="linux" {
    if f,err:=os.Open("/proc/cpuinfo"); err==nil { defer f.Close(); sc:=bufio.NewScanner(f); for sc.Scan(){ line:=sc.Text(); if strings.HasPrefix(line,"model name")||strings.HasPrefix(line,"Hardware"){ p:=strings.SplitN(line,":",2); if len(p)==2{return strings.TrimSpace(p[1])}}}}
  }
  if runtime.GOOS=="darwin" {
    if out,err:=exec.Command("sysctl","-n","machdep.cpu.brand_string").Output(); err==nil{return strings.TrimSpace(string(out))}
    if out,err:=exec.Command("sysctl","-n","hw.model").Output(); err==nil{return strings.TrimSpace(string(out))}
  }
  if runtime.GOOS=="windows" {
    if out,err:=exec.Command("powershell","-NoProfile","-Command","(Get-CimInstance Win32_Processor | Select-Object -First 1 -ExpandProperty Name)").Output(); err==nil{return strings.TrimSpace(string(out))}
  }
  return runtime.GOARCH
}
func detectRAM()(float64,float64){
  if runtime.GOOS=="linux" {
    if b,err:=os.ReadFile("/proc/meminfo");err==nil{var t,a float64;for _,line:=range strings.Split(string(b),"\n"){var kb float64;if _,e:=fmt.Sscanf(line,"MemTotal: %f kB",&kb);e==nil{t=kb};if _,e:=fmt.Sscanf(line,"MemAvailable: %f kB",&kb);e==nil{a=kb}};return t/1024/1024,a/1024/1024}
  }
  if runtime.GOOS=="darwin" {if out,err:=exec.Command("sysctl","-n","hw.memsize").Output();err==nil{n,_:=strconv.ParseFloat(strings.TrimSpace(string(out)),64);t:=n/1024/1024/1024;return t,t*0.75}}
  if runtime.GOOS=="windows" {cmd:=exec.Command("powershell","-NoProfile","-Command",`$o=Get-CimInstance Win32_OperatingSystem; "$($o.TotalVisibleMemorySize),$($o.FreePhysicalMemory)"`);if out,err:=cmd.Output();err==nil{p:=strings.Split(strings.TrimSpace(string(out)),",");if len(p)==2{t,_:=strconv.ParseFloat(p[0],64);f,_:=strconv.ParseFloat(p[1],64);return t/1024/1024,f/1024/1024}}}
  return 0,0
}
func detectGPUs(totalRAM float64)[]GPU{
  out:=detectNVIDIA()
  if runtime.GOOS=="darwin"&&runtime.GOARCH=="arm64"{name:="Apple Silicon GPU";if b,err:=exec.Command("system_profiler","SPDisplaysDataType").Output();err==nil{for _,line:=range strings.Split(string(b),"\n"){if strings.Contains(line,"Chipset Model:"){p:=strings.SplitN(line,":",2);if len(p)==2{name=strings.TrimSpace(p[1]);break}}}};out=append(out,GPU{Name:name,VRAMGB:totalRAM,FreeVRAMGB:totalRAM*0.75,Backend:"metal",Unified:true})}
  return out
}
func detectNVIDIA()[]GPU{
  b,err:=exec.Command("nvidia-smi","--query-gpu=name,memory.total,memory.free","--format=csv,noheader,nounits").Output();if err!=nil{return nil}
  var out []GPU;for _,line:=range bytes.Split(bytes.TrimSpace(b),[]byte("\n")){p:=strings.Split(string(line),",");if len(p)<3{continue};t,_:=strconv.ParseFloat(strings.TrimSpace(p[1]),64);f,_:=strconv.ParseFloat(strings.TrimSpace(p[2]),64);out=append(out,GPU{Name:strings.TrimSpace(p[0]),VRAMGB:t/1024,FreeVRAMGB:f/1024,Backend:"cuda"})};return out
}
func CommandExists(name string)bool{_,err:=exec.LookPath(name);return err==nil}
