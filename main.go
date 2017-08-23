package main

import (
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"runtime"
	"io/ioutil"
	"strings"
	"encoding/json"
	"fmt"
	"strconv"
	"os"
	"bufio"
	"io"
	"errors"
)

func getcurdir() (string, error) {
	//get absolute path, which will be used to locate html and js files
	_, filename, _, _ := runtime.Caller(1)
	dir, err := filepath.Abs(filepath.Dir(filename))
	if err != nil {
		return "", err
	}
	return dir + "/", nil
}

func main() {
	modelCache := make(map[string]*Segmentation)

	curDir, _ := getcurdir()
	g := gin.Default()

	g.StaticFS("/", http.Dir(curDir+"frontend/dist"))

	g.OPTIONS("/segmentation", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
	})
	g.POST("/segmentation", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		payload, _ := ioutil.ReadAll(c.Request.Body)
		req := make(map[string]string)
		json.Unmarshal(payload, &req)

		fmt.Println("Get request:", req)

		model := req["model-file"]
		dict := req["dict-file"]
		batch_size, _ := strconv.Atoi(req["batch-size"])
		lines := strings.Split(req["strings"], "\n")

		var seg *Segmentation
		var ok bool
		if seg, ok = modelCache[model + dict]; !ok {

			var err error
			seg, err = NewSegmentation(model, dict, batch_size)
			modelCache[model + dict] = seg
			if err != nil {
				c.Error(err)
			}
		}

		res, err := seg.SegmentLine(lines)
		if err != nil {
			c.Error(err)
		}

		c.JSON(http.StatusOK, res)
	})

	g.Run(":8888")
}

type Segmentation struct {
	session *tf.Session
	tfgraph *tf.Graph

	char2id map[string]int32
	id2char map[int32]string

	batch_size int
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

func buildDict(dict_filename string) (char2id map[string]int32, id2char map[int32]string, err error) {
	/*
	   build char2id and id2char dict by dictfile, which contains nearly all symbols in chinese
	*/
	fin, err := os.Open(dict_filename)
	if err != nil {
		return
	}
	defer fin.Close()

	char2id = map[string]int32{}
	id2char = map[int32]string{}

	char2id["<bos>"] = int32(0)
	char2id["<eos>"] = int32(1)
	char2id["<unk>"] = int32(2)
	id2char[0] = "<bos>"
	id2char[1] = "<eos>"
	id2char[2] = "<unk>"

	rd := bufio.NewReader(fin)
	idx := int32(3)
	for {
		line, err := rd.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		// remove \n
		line = strings.Replace(line, "\n", "", -1)
		_, ok := char2id[line]
		if ok == false {
			char2id[line] = idx
			id2char[idx] = line
			idx += 1
		}
	}
	return
}

func (s *Segmentation) SegmentLine(lines []string) (results [][]string, err error) {
	var (
		idxs_b = make([][]int32, s.batch_size)
		length = make([]int32, s.batch_size)
		counter = 0
	)
	originLen := len(lines)
	if len(lines) % s.batch_size != 0 {
		reqNum := s.batch_size - (len(lines) % s.batch_size)
		for i := 0; i < reqNum; i++ {
			lines = append(lines, lines[originLen - 1])
		}
	}

	results = make([][]string, len(lines))
	segInterResult := make([][]int32, len(lines))

	for idx, line := range lines {
		inferText := strings.Replace(line, " ", "", -1)
		idxs := s.Text2Idx(inferText, s.char2id, 100)
		idxs_b[counter] = idxs
		length[counter] = int32(len([]rune(inferText)))
		counter++

		if counter == s.batch_size {
			tensor, err1 := tf.NewTensor(idxs_b)
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
					s.tfgraph.Operation("model/batch_in").Output(0):  tensor,
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
				segInterResult[idx - s.batch_size + 1 + i] = seg
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

func NewSegmentation(modelfile, dict_filename string, batch_size int) (seg *Segmentation, err error) {
	seg = &Segmentation{
	}
	// Build a chinese dict for character to index
	seg.char2id, seg.id2char, err = buildDict(dict_filename)

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

	seg.batch_size = batch_size
	return
}

func parseSegmentatoinResult(seg []int32, str []rune) (res []string, err error) {
	// accepts batches of text data as input. batch_size == 1
	// output[0].Value() contains the segment information
	// tag_map{0:B, 1:M, 2:E, 3:S }
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
