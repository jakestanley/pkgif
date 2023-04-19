<template>
  <video id="preview" width="320" ref="video" controls @pause="updatePaused">
    <source id="preview_source" ref="videoSrc" type="video/mp4" />
  </video>
  <br />
  <input v-model="clipStart" type="text" placeholder="0" />
  <button v-on:click="setClipStartToHere">from here</button>
  <button v-on:click="goToClipStart()">Go to</button>
  <br />
  <input v-model="clipEnd" type="text" placeholder="infinity" />
  <button v-on:click="setClipEndToHere">to here</button>
  <button v-on:click="goToClipEnd()">Go to</button>
  <br />
  <button v-on:click="createClip()">Create clip</button>
</template>

<script lang="ts">
import axios from 'axios'
import { ref } from 'vue'

export default {
  data() {
    return {
      id: this.$route.params.id,
      clipStart: this.$route.query.clipStart,
      clipEnd: null
    }
  },
  mounted() {
    // TODO while video not downloaded, spin and keep spinning, then update videoSrc.src?
    this.$refs.videoSrc.src = 'http://localhost:7131/video/' + this.$data.id + '/preview'
    axios.get('http://localhost:7131/video/' + this.id).then((response) => {
      console.log(response)
      console.log(response.data)
      this.$data.clipEnd = response.data.length
    })
  },
  methods: {
    setClipStartToHere() {
      this.clipStart = this.$refs.video.currentTime
    },
    setClipEndToHere() {
      this.clipEnd = this.$refs.video.currentTime
    },
    goToClipStart() {
      let video = this.$refs.video
      if (this.$refs.video.currentTime == this.$data.clipStart) {
        video.play()
      } else {
        this.$refs.video.currentTime = this.$data.clipStart
        video.pause()
      }
    },
    goToClipEnd() {
      let video = this.$refs.video
      if (this.$refs.video.currentTime == this.$data.clipEnd) {
        video.play()
      } else {
        this.$refs.video.currentTime = this.$data.clipEnd
        video.pause()
      }
    },
    createClip() {
      axios
        .post('http://localhost:7131/clip', {
          videoId: this.id,
          clipStart: Number(this.clipStart),
          clipEnd: Number(this.clipEnd)
        })
        .then((response) => {
          this.$router.push({
            name: 'clip',
            params: {
              id: response.data.id
            }
          })
        })
    }
  }
}
</script>
