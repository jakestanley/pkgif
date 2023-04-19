<template>
  <h1>Clip</h1>

  <video id="preview" width="320" ref="video" controls @pause="updatePaused">
    <source id="preview_source" ref="videoSrc" type="video/mp4" />
  </video>
  <br />

  <input type="text" disabled="true" :value="paused" />

  <h1>Captions</h1>
  <button v-on:click="addCaption()">Add caption</button>
  <div v-show="captions.length > 0">
    <table>
      <thead>
        <th>#</th>
        <th>Start</th>
        <th>Caption</th>
        <th>End</th>
        <th></th>
      </thead>
      <tbody>
        <tr v-for="(caption, i) in captions">
          <td>{{ i }}</td>
          <td>
            <input type="text" @change="inputCaptionStart($event, i)" v-model="caption.start" />
          </td>
          <td><input type="text" v-model="caption.text" /></td>
          <td><input type="text" @change="inputCaptionEnd($event, i)" v-model="caption.end" /></td>
          <td>
            <button @click="scrubTo(i)">Go to</button>
            <button @click="setCaptionStart(i)">Start</button>
            <button @click="setCaptionEnd(i)">End</button>
            <button @click="removeCaption(i)">Remove</button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
  <button v-on:click="save()">Save</button>
  <button v-on:click="preview()">Preview</button>
  <h1>Render</h1>
  <button v-on:click="render()">Render</button>
</template>

<script lang="ts">
import axios from 'axios'
import { ref } from 'vue'

export default {
  data() {
    return {
      id: this.$route.params.id,
      paused: 0,
      captions: []
    }
  },
  mounted() {
    this.$refs.videoSrc.src = 'http://localhost:7131/clip/' + this.$data.id + '/preview'
    axios.get('http://localhost:7131/clip/' + this.id).then((response) => {
      console.log(response)
      console.log(response.data)
      this.clipEnd = response.data.videoLength
      this.captions = response.data.captions
    })
  },
  methods: {
    updatePaused(event) {
      this.$data.paused = event.target.currentTime
    },
    addCaption() {
      let video = this.$refs.video
      let caption = {
        start: Number(video.currentTime)
      }
      this.$data.captions.push(caption)
    },
    scrubTo(i) {
      let video = this.$refs.video
      // clicking and clicking again will resume play from this point
      if (this.$data.captions[i].start == video.currentTime) {
        video.play()
      } else {
        video.currentTime = this.$data.captions[i].start
        video.pause()
      }
    },
    setCaptionStart(i) {
      this.$data.captions[i].start = this.$refs.video.currentTime
    },
    setCaptionEnd(i) {
      this.$data.captions[i].end = this.$refs.video.currentTime
    },
    inputCaptionStart(event, i) {
      this.$data.captions[i].start = Number(event.target.value)
    },
    inputCaptionEnd(event, i) {
      this.$data.captions[i].end = Number(event.target.value)
    },
    removeCaption(i) {
      let newCaptions = []
      for (let index = 0; index < this.$data.captions.length; index++) {
        if (index != i) {
          newCaptions.push(this.$data.captions[index])
        }
      }
      // TODO sort by start time
      this.$data.captions = newCaptions
    },
    save() {
      axios
        .put('http://localhost:7131/clip/' + this.id, {
          captions: this.$data.captions
        })
        .then((response) => {
          console.log('got put response')
        })
        .catch((err) => {
          console.debug(err)
        })
    },
    preview() {
      // TODO set loading and set not loading in response handler
      axios.get('http://localhost:7131/clip/' + this.id + '/preview').then((response) => {
        this.$refs.video.load()
        console.log('preview ready')
      })
    },
    render() {
      // TODO router for configuring render options
      // TODO
      // axios.put()
    }
  }
}
</script>
