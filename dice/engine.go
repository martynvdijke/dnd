package dice

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"

	"github.com/dop251/goja"
)

// Pool is a reusable dice engine pool (not goroutine-safe per instance,
// but the pool mediates access). Each engine has its own goja runtime.
type Pool struct {
	mu   sync.Mutex
	free []*Engine
}

// Engine wraps a goja runtime with the dice-roller library loaded.
type Engine struct {
	vm      *goja.Runtime
	rollFn  goja.Callable
	pool    *Pool
}

// NewPool creates a pool of n engines. Each engine pre-loads the dice-roller bundle.
func NewPool(n int) (*Pool, error) {
	p := &Pool{}
	for range n {
		eng, err := newEngine(p)
		if err != nil {
			return nil, fmt.Errorf("dice engine init: %w", err)
		}
		p.free = append(p.free, eng)
	}
	return p, nil
}

// Acquire gets an engine from the pool. Call Release when done.
func (p *Pool) Acquire() *Engine {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.free) == 0 {
		// Grow pool on demand
		eng, err := newEngine(p)
		if err != nil {
			return nil
		}
		return eng
	}
	eng := p.free[len(p.free)-1]
	p.free = p.free[:len(p.free)-1]
	return eng
}

// Release returns an engine to the pool.
func (p *Pool) Release(eng *Engine) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.free = append(p.free, eng)
}

// RollResult maps the JSON output of DiceRoll.toJSON()
type RollResult struct {
	Notation   string          `json:"notation"`
	Total      json.Number     `json:"total"`
	Rolls      json.RawMessage `json:"rolls,omitempty"`
	Output     string          `json:"output"`
	Error      string          `json:"error,omitempty"`
}

// Roll evaluates a dice expression and returns the result.
func (p *Pool) Roll(expression string) (*RollResult, error) {
	eng := p.Acquire()
	if eng == nil {
		return nil, fmt.Errorf("no available dice engine")
	}
	defer p.Release(eng)

	return eng.roll(expression)
}

// newEngine creates a single dice engine with its own goja runtime.
func newEngine(pool *Pool) (*Engine, error) {
	vm := goja.New()

	// Provide require("crypto") polyfill with crypto/rand
	cryptoObj := vm.NewObject()
	_ = cryptoObj.Set("randomBytes", func(call goja.FunctionCall) goja.Value {
		size := int(call.Argument(0).ToInteger())
		if size <= 0 {
			return vm.ToValue(nil)
		}
		buf := make([]byte, size)
		_, err := rand.Read(buf)
		if err != nil {
			panic(vm.NewTypeError(err.Error()))
		}
		// Return as ArrayBuffer (what random-js expects from crypto.randomBytes)
		ab := vm.NewArrayBuffer(buf)
		return vm.ToValue(ab)
	})
	_ = vm.Set("crypto", cryptoObj)

	// Override __require to return our crypto module
	_ = vm.Set("require", func(call goja.FunctionCall) goja.Value {
		mod := call.Argument(0).String()
		if mod == "crypto" {
			return cryptoObj
		}
		panic(vm.NewTypeError("require('%s') not available", mod))
	})

	// Also provide globalThis.crypto for browser-like access if needed
	_ = vm.Set("globalThis", vm.GlobalObject())

	// Evaluate the bundled dice-roller
	bundleJS := BundleJS // from bundle_stub.go
	_, err := vm.RunString(bundleJS)
	if err != nil {
		return nil, fmt.Errorf("eval bundle: %w", err)
	}

	// Get the __diceRoll function
	rollVal := vm.Get("__diceRoll")
	if rollVal == nil || goja.IsUndefined(rollVal) {
		return nil, fmt.Errorf("__diceRoll not found in bundle")
	}
	rollFn, ok := goja.AssertFunction(rollVal)
	if !ok {
		return nil, fmt.Errorf("__diceRoll is not a function")
	}

	return &Engine{
		vm:     vm,
		rollFn: rollFn,
		pool:   pool,
	}, nil
}

// roll evaluates an expression in this engine's runtime.
func (e *Engine) roll(expression string) (*RollResult, error) {
	resultVal, err := e.rollFn(goja.Undefined(), e.vm.ToValue(expression))
	if err != nil {
		return nil, fmt.Errorf("js roll error: %w", err)
	}

	resultStr := resultVal.String()

	var result RollResult
	if err := json.Unmarshal([]byte(resultStr), &result); err != nil {
		return nil, fmt.Errorf("parse result: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("dice parse error: %s", result.Error)
	}

	return &result, nil
}

// Helper to read crypto/rand bytes as uint32
func cryptoUint32() uint32 {
	var buf [4]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		panic(err)
	}
	return binary.BigEndian.Uint32(buf[:])
}

// Helper for crypto/rand int
func cryptoInt(max int64) int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		panic(err)
	}
	return n.Int64()
}
