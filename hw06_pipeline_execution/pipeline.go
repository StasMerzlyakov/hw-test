package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func wrapChannelWithDone(in In, done In) Out {
	out := make(chan any, 1) // buffered channel
	go func() {
		defer close(out)
		defer func() {
			go func() {
				for { // read, if any, to terminate goroutines writing to the channel
					if _, ok := <-in; !ok {
						break
					}
				}
			}()
		}()
		var ok bool
		var val any

		inWrap := in
		var outWrap chan any // nil

		for {
			select {
			case <-done:
				return
			case val, ok = <-inWrap:
				if !ok {
					return
				}
				inWrap = nil
				outWrap = out
			case outWrap <- val:
				outWrap = nil
				inWrap = in
			}
		}
	}()

	return out
}

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	prevCh := wrapChannelWithDone(in, done)

	for _, stage := range stages {
		prevCh = stage(prevCh) //  stage[i] out == stage[i+1] in
		prevCh = wrapChannelWithDone(prevCh, done)
	}

	return prevCh
}
