// Copyright 2025 NAEOS contributors
// SPDX-License-Identifier: Apache-2.0

package controlplane

import (
 "context"
 "errors"
 "sync"
 "testing"
)

type memoryReplayGuard struct { mu sync.Mutex; claims map[string]bool; err error }
func (g *memoryReplayGuard) Claim(_ context.Context, scope,id string)(bool,error) {
 g.mu.Lock(); defer g.mu.Unlock()
 if g.err!=nil { return false,g.err }
 key:=scope+":"+id
 if g.claims[key] { return false,nil }
 g.claims[key]=true
 return true,nil
}
func TestReplayGuardConcurrentClaimOnlyOneWinner(t *testing.T) {
 g:=&memoryReplayGuard{claims:map[string]bool{}}
 const workers=48
 var wg sync.WaitGroup
 var winners int
 var mu sync.Mutex
 for i:=0;i<workers;i++ { wg.Add(1); go func(){ defer wg.Done(); ok,err:=g.Claim(context.Background(),"scope","aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"); if err!=nil { t.Errorf("Claim error: %v",err); return }; if ok { mu.Lock(); winners++; mu.Unlock() } }() }
 wg.Wait()
 if winners!=1 { t.Fatalf("winners=%d; want exactly one",winners) }
}
func TestReplayGuardFailClosedOnStorageError(t *testing.T) {
 g:=&memoryReplayGuard{claims:map[string]bool{},err:errors.New("db unavailable")}
 ok,err:=g.Claim(context.Background(),"scope","aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
 if ok || err==nil { t.Fatalf("claim=%v err=%v; expected fail closed",ok,err) }
}
