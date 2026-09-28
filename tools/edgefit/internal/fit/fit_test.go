package fit

import (
  "testing"
  "github.com/hippoley/ClawManager/tools/edgefit/internal/catalog"
  "github.com/hippoley/ClawManager/tools/edgefit/internal/hardware"
)
func TestPlanFitsSmallModelOn24GBGPU(t *testing.T){spec:=hardware.Specs{AvailableRAMGB:32,CPUCores:16,GPUs:[]hardware.GPU{{Name:"test",VRAMGB:24,FreeVRAMGB:22,Backend:"cuda"}}};m:=catalog.Model{ID:"test-7b",ParamsB:7,ActiveB:7,Context:32768,Quality:70,Coding:80,Reasoning:75,Formats:[]string{"gguf"}};r:=Plan(spec,m,"Q4_K_M","coding",8192);if r.Fit=="too-tight"{t.Fatalf("expected fit, got %+v",r)};if r.RunPath!="gpu"{t.Fatalf("expected gpu path, got %s",r.RunPath)};if r.EstimatedTPS<=0{t.Fatal("expected positive tps")}}
func TestTooLargeModelDoesNotPretendToFit(t *testing.T){spec:=hardware.Specs{AvailableRAMGB:8,CPUCores:4};m:=catalog.Model{ID:"huge",ParamsB:70,ActiveB:70,Context:32768,Quality:90};r:=Plan(spec,m,"Q8_0","general",8192);if r.Fit!="too-tight"{t.Fatalf("expected too-tight, got %+v",r)}}
