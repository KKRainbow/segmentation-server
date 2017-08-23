<template>
  <div id="app">
    <index></index>
  </div>
</template>

<script>
  import Index from './components/Index'
  import Vue from 'vue'
  export default {
    name: 'app',
    components: {
      Index
    }
  }
</script>

<style>
  * {
    margin: 0;
    padding: 0;
  }

  #app {
    font-family: 'Avenir', Helvetica, Arial, sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    text-align: center;
    color: #2c3e50;
    margin-top: 60px;

    width: 100%;
    height: 100%;
  }

</style>
