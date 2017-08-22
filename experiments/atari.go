package experiments

import (
	"github.com/unixpickle/anyrl"
	"github.com/unixpickle/anyvec"
	"github.com/unixpickle/essentials"
	gym "github.com/unixpickle/gym-socket-api/binding-go"
)

const (
	atariWidth  = 80
	atariHeight = 105
	atariScale  = 2
)

var atariActionSizes = map[string]int{
	"Pong-v0":     6,
	"Breakout-v0": 4,
}

type atariEnv struct {
	Env    anyrl.Env
	Closer gym.Env
}

func newAtariEnvs(c anyvec.Creator, g *GameFlags, n int) ([]Env, error) {
	var res []Env
	for i := 0; i < n; i++ {
		client, err := gym.Make(g.GymHost, g.Name)
		if err != nil {
			CloseEnvs(res)
			return nil, err
		}
		env, err := anyrl.GymEnv(c, client, false)
		if err != nil {
			CloseEnvs(res)
			return nil, err
		}
		res = append(res, &atariEnv{Env: env, Closer: client})
	}
	return res, nil
}

func (a *atariEnv) Reset() (obs anyvec.Vector, err error) {
	obs, err = a.Env.Reset()
	if err != nil {
		return
	}
	obs = downsampleAtariObs(obs)
	return
}

func (a *atariEnv) Step(action anyvec.Vector) (obs anyvec.Vector, reward float64,
	done bool, err error) {
	obs, reward, done, err = a.Env.Step(action)
	if err != nil {
		return
	}
	obs = downsampleAtariObs(obs)
	return
}

func (a *atariEnv) Close() error {
	return a.Closer.Close()
}

func downsampleAtariObs(obs anyvec.Vector) anyvec.Vector {
	comps := obsComponents(obs)
	newComps := make([]float64, 0, atariWidth*atariHeight)
	for y := 0; y < atariHeight; y++ {
		for x := 0; x < atariWidth; x++ {
			idx := 3 * atariScale * (y*atariWidth*atariScale + x)
			var sum float64
			for z := 0; z < 3; z++ {
				sum += comps[idx+z]
			}
			val := essentials.Round(sum / 3)
			newComps = append(newComps, val)
		}
	}
	return obs.Creator().MakeVectorData(obs.Creator().MakeNumericList(newComps))
}

func obsComponents(obs anyvec.Vector) []float64 {
	switch data := obs.Data().(type) {
	case []float64:
		return data
	case []float32:
		res := make([]float64, len(data))
		for i, x := range data {
			res[i] = float64(x)
		}
		return res
	default:
		panic("unsupported numeric type")
	}
}