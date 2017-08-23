<template>
  <div id="container" >
    <el-alert v-show="errorMsg.length != 0"
      :title="errorMsg"
      type="error">
    </el-alert>
    <el-row>
      分词
    </el-row>

    <el-row type="flex" justify="center">
      <el-col :span="20">
        <el-input placeholder="请输入Model File的绝对路径" v-model="params.modelFile">
          <template slot="prepend">Model File</template>
        </el-input>
      </el-col>
    </el-row>

    <el-row type="flex" justify="center">
      <el-col :span="20">
        <el-input placeholder="请输入Dict File的绝对路径" v-model="params.dict">
          <template slot="prepend">Dict File</template>
        </el-input>
      </el-col>
    </el-row>

    <el-row type="flex" justify="start">
      <el-col :span="2"></el-col>
      <el-col :span="8">
        <el-input placeholder="请输入Batch Size" v-model="params.batch">
          <template slot="prepend">Batch Size</template>
        </el-input>
      </el-col>
    </el-row>

    <el-row type="flex" justify="center">
      <el-col :span="20">
        <div>
          <el-form>
            <el-input
              type="textarea"
              :autosize="{ minRows: 4, maxRows: 16}"
              placeholder="请输入要分词的句子，两个句子用回车隔开"
              v-model="params.lines">
            </el-input>
          </el-form>
        </div>
      </el-col>
    </el-row>

    <el-row type="flex" justify="start">
      <el-col :span="2"></el-col>
      <el-col :span="0">
        <el-button type="primary" size="large" @click="submit">提交</el-button>
      </el-col>
    </el-row>

    <el-row type="flex" justify="center">
      <el-col :span="20" v-loading.body="loading">
        <el-tabs v-model="activeLogTabName" type="border-card" editable @edit="logsOperationHandler" >
          <el-tab-pane
            :key="item.name"
            v-for="(item, index) in logs"
            :label="item.title"
            :name="item.name"
          >
            <el-table
              :data="item.segments"
              align="left"
              style="width: 100%">
              <el-table-column
                prop="str"
              >
              </el-table-column>
            </el-table>
            <el-row type="flex" justify="start" v-for="item in description.split('\n')" style="margin-top: 20px">
                <p>{{item}}</p>
            </el-row>
          </el-tab-pane>
        </el-tabs>
      </el-col>
    </el-row>


  </div>
</template>

<script>
  const MAX_LOG = 12;
  import Axios from 'axios';
  import ElCol from "element-ui/packages/col/src/col";
  export default {
    components: {ElCol},
    name: 'index',
    data () {
      return {
        params: {
          modelFile: "",
          dict: "",
          batch: "",
          lines: "",
        },

        activeLogTabName: "1",

        logs: [
        ],

        errorMsg: "",
        infoMsg: "fdsa",

        loading: false,
      }
    },

    methods: {

      saveOptions() {
        localStorage.setItem("params", JSON.stringify(this.params));
      },

      loadOptions() {
        this.params = JSON.parse(localStorage.getItem("params"));
      },

      loadLogs() {
        this.logs = JSON.parse(localStorage.getItem("logs"));
        this.activeLogTabName = this.logs[0].name;
      },

      saveLogs() {
        localStorage.setItem("logs", JSON.stringify(this.logs));
      },

      submit() {
        this.loading = true;
        this.saveOptions();
        this.params.lines = this.params.lines.split("\n").filter((item) => {return item.length !== 0}).join("\n");
        let data = {
          'model-file': this.params.modelFile,
          'dict-file' : this.params.dict,
          'batch-size' : this.params.batch,
          'strings' : this.params.lines,
        };

        Axios.post(API_URL + 'segmentation', data).then((res) => {
          let lines = res.data.map((item) => {
            item = item.filter((item) => {return item.trim().length !== 0});
            return item.join(" | ")
          }).map((item) => {return {str: item}});
          /*
              lines should be format of following:
              lines: [
                {
                  str: "kfjldsjalkfdjl|lkfsdjalk"
                },
                {
                  str: "kfjldsjalkfdjl|lkfsdjalk"
                }
              ]
          */

          if (this.logs.length >= MAX_LOG) {
            this.logs.splice(this.logs.length - 1, 1);
          }

          this.logs.unshift({
            params: JSON.stringify(this.params),
            title: new Date().toLocaleTimeString(),
            name: Date.now().toString(),
            segments: lines,
          });

          this.activeLogTabName = this.logs[0].name;

          this.saveLogs();
          this.loading = false;
        }).catch((err) => {
          this.errorMsg = err.toString();
          setTimeout(() => {
            this.errorMsg = "";
          }, 5000);
          this.loading = false;
        });

      },

      getLogIndex(name) {
        let idx = 0;
        for (let tmp of this.logs) {
          if (tmp.name === name) {
            return idx
          }
          idx++;
        }
      },

      getCurrentLogIndex() {
        return this.getLogIndex(this.activeLogTabName);
      },

      logsOperationHandler(targetName, action) {
        if (action === 'remove') {
          let isCur = targetName === this.activeLogTabName;
          let idx = this.getLogIndex(targetName);
          this.logs.splice(idx, 1);
          if (isCur) {
            this.activeLogTabName = this.logs[0].name;
          }
        }
      }
    },

    mounted() {
      this.loadOptions();
      this.loadLogs();
    },

    computed: {
      description() {
        let tmp = this.logs[this.getCurrentLogIndex()];
        let params = JSON.parse(tmp.params);
        let d = new Date(parseInt(tmp.name));
        return `Model File: ${params.modelFile} \n Dict File: ${params.dict} \n Time: ${d.toLocaleDateString()} ${d.toLocaleTimeString()}`
      }
    }
  }
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style lang="scss" scoped>
  #container {
    .el-row {
      margin-bottom: 20px;
    }

    .el-alert {
      position: fixed;
      top: 0;
      z-index: 99;
    }

    .segment {
      background-color: #f9fafc;
    }
  }
</style>
