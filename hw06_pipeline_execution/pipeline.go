package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	out := make(chan interface{})

	var prevCh = in
	for _, stage := range stages {
		prevCh = stage(prevCh) //  stage[i] out == stage[i+1] in
	}

	go func() {
		for {
			select {
			case <-done:
				return
			case val := <-prevCh: // read stage[len(stages)-1] result and sent to out
				out <- val
			}
		}
	}()

	return out

}
