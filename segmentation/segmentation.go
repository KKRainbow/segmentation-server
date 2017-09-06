package segmentation

import (
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"strings"
	"errors"
	"strconv"
	"io/ioutil"
)

type Segmentation struct {
	session *tf.Session
	tfgraph *tf.Graph

	dagBuilder *DAGBuilder

	maxLength int
	batchSize int
}

func (*Segmentation) Text2Idx(text string, char2id map[string]int32, max_length int) (idxs []int32) {
	/*
	   take a string and return idx of it by char2id map
	   if the final idx array excludes max_length, return emtpy array
	*/
	for _, v := range text {
		key := string(v)
		idx, ok := char2id[key]
		if ok {
			idxs = append(idxs, idx)
		} else {
			idxs = append(idxs, char2id["<unk>"])
		}
	}
	if len(idxs) < max_length {
		pad_num := max_length - len(idxs)
		for i := 0; i < pad_num; i++ {
			idxs = append(idxs, char2id["<eos>"])
		}
	} else {
		// if exclude max_length, return emtpy array
		idxs = idxs[:0]
	}
	return idxs
}

func (s *Segmentation) SegmentLine(lines []string) (results [][]string, err error) {
	var (
		idxs_b_fd = make([][][]int32, s.batchSize)
		idxs_b_bd = make([][][]int32, s.batchSize)
		length = make([]int32, s.batchSize)
		counter = 0
	)
	originLen := len(lines)
	if len(lines) % s.batchSize != 0 {
		reqNum := s.batchSize - (len(lines) % s.batchSize)
		for i := 0; i < reqNum; i++ {
			lines = append(lines, lines[originLen - 1])
		}
	}

	results = make([][]string, len(lines))
	segInterResult := make([][]int32, len(lines))

	for idx, line := range lines {
		inferText := strings.Replace(line, " ", "", -1)
		idxs_b_fd[counter] = s.dagBuilder.buildMatrix([]rune(inferText), s.maxLength,true)
		idxs_b_bd[counter] = s.dagBuilder.buildMatrix([]rune(inferText), s.maxLength,false)
		length[counter] = int32(len([]rune(inferText)))
		counter++

		if counter == s.batchSize {
			tensor_fd, err1 := tf.NewTensor(idxs_b_fd)
			tensor_bd, err1 := tf.NewTensor(idxs_b_bd)
			if err1 != nil {
				err = err1
				return
			}
			len_tensor, err1 := tf.NewTensor(length)
			if err1 != nil {
				err = err1
				return
			}

			o, err1 := s.session.Run(
				map[tf.Output]*tf.Tensor{
					s.tfgraph.Operation("model/batch_fd_dag").Output(0):  tensor_fd,
					s.tfgraph.Operation("model/batch_bd_dag").Output(0):  tensor_bd,
					s.tfgraph.Operation("model/batch_len").Output(0): len_tensor,
				},
				[]tf.Output{
					s.tfgraph.Operation("model/seg_pred").Output(0),
				},
				nil)
			if err1 != nil {
				err = err1
				return
			}

			segs := o[0].Value().([][]int32)
			for i, seg := range segs {
				segInterResult[idx - s.batchSize + 1 + i] = seg
			}

			counter = 0
		}
	}

	if counter != 0 {
		err = errors.New("Counter should be zero but " + strconv.Itoa(counter))
	}

	for i, seg := range segInterResult {
		s, _ := parseSegmentatoinResult(seg, []rune(lines[i]))
		results[i] = s
	}
	results = results[:originLen]
	return
}

func NewSegmentation(modelfile, dict_filename string, maxLength, batchSize int, dagBuilder *DAGBuilder) (seg *Segmentation, err error) {
	seg = &Segmentation{
	}
	// Build a chinese dict for character to index
	// Load the serialized GraphDef from a file.
	model, err := ioutil.ReadFile(modelfile)
	if err != nil {
		return
	}

	// Construct an in-memory graph from the serialized from.
	seg.tfgraph = tf.NewGraph()
	if err = seg.tfgraph.Import(model, ""); err != nil {
		return
	}

	// Create a session for inference over graph.
	seg.session, err = tf.NewSession(seg.tfgraph, nil)
	if err != nil {
		return
	}

	seg.batchSize = batchSize
	seg.maxLength = maxLength

	seg.dagBuilder = dagBuilder
	return
}

func parseSegmentatoinResult(seg []int32, str []rune) (res []string, err error) {
	// accepts batches of text data as input. batchSize == 1
	// output[0].Value() contains the segment information
	// tag_map{0:B, 1:M, 2:E, 3:S }
	seg = seg[1:]
	if len(str) == 0 {
		return nil, errors.New("empty str")
	}
	if len(seg) < len(str) {
		return nil, errors.New("len of seg less than of str")
	}
	res = make([]string, 0)
	beg := 0
	err = nil
	for i := range str {
		switch seg[i] {
		case 0:
			if i > beg {
				res = append(res, string(str[beg:i]))
			}
			beg = i
		case 1:
			continue
		case 2:
			if i > beg {
				res = append(res, string(str[beg:i+1]))
			}
			beg = i + 1
		case 3:
			res = append(res, string(str[i:i+1]))
			beg = i + 1
		}
	}
	res = append(res, string(str[beg:]))
	return
}
