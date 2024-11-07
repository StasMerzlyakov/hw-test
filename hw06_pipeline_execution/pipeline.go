package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	out := make(chan interface{})

	startCh := make(chan interface{})
	var prevCh In

	prevCh = startCh
	for _, stage := range stages {
		prevCh = stage(prevCh) //  stage[i] out == stage[i+1] in

	}

	go func() {
		defer close(out)

		for {
			select {
			case <-done:
				close(startCh)
				done = nil
			case vl, ok := <-prevCh:
				if !ok {
					return
				} else {
					startCh <- vl
				}
			}

		}
	}()

	/*go func() {
		pCh := prevCh
		var outCh Bi
		var val interface{}
		var dnCh = done
		var ok bool
		for {
			select {
			case <-dnCh:
				close(startCh)
				dnCh = nil // read once

			case val, ok = <-pCh: // read stage[len(stages)-1] result and sent to out
				if ok {
					fmt.Printf("received result %v\n", val)
					pCh = nil
					outCh = out
				} else {
					close(out)
					return
				}
			case outCh <- val:
				fmt.Printf("send result %v\n", val)
				outCh = nil
				pCh = prevCh
			}
		}
	}() */

	return out

}
