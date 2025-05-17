/* 
	This packages defines jobs meaning the tasks which run in background 
  i.e concurrently which are to be processed by the workers 
*/
package worklist

type Entry struct {
	Path string
}

// Stores the path of the files which are to be searched 
type Worklist struct {
	jobs chan Entry
}

func (w *Worklist) Add(work Entry) {
	w.jobs <- work
}

func (w *Worklist) Next() Entry {
	j := <- w.jobs
	return j
}

func New(bufSize int) Worklist {
	return Worklist{ make(chan Entry, bufSize)}
}

func NewJob(path string) Entry {
	return Entry{ path }
}

// Terminate workers by passing mpty path to each
func (w *Worklist) Finalize(numWorkers int){
	// iterates over numworkers
	for range numWorkers{
		w.Add(Entry{""})
	}
}
