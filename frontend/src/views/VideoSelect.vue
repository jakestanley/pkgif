<template>
  <h1>Select a video</h1>

  <input
    v-model="location"
    id="location_computer"
    type="radio"
    name="location"
    value="computer"
  /><label for="location_computer"> From my computer</label><br />
  <input
    v-model="location"
    id="location_internet"
    type="radio"
    name="location"
    value="internet"
  /><label for="location_internet"> From the Internet</label>

  <!-- <div v-show="location=='computer'">
        <label for="file">File </label><input type="file" id="file" v-on:change="updateComputerPath"/>
    </div> -->

  <div v-show="location == 'internet'">
    <p>e.g https://www.youtube.com/watch?v=dQw4w9WgXcQ</p>
    <label for="url">URL </label
    ><input
      type="text"
      v-on:change="getVideoInfo()"
      placeholder="https://www.youtube.com/watch?v=dQw4w9WgXcQ"
      id="url"
      v-model="videoUrl"
    />
  </div>

  <em v-text="title"></em>
  <img :src="thumbnailUrl" width="168" />

  <div>
    <button :disabled="videoId == null" v-on:click="createVideo()">Load</button>
  </div>
</template>

<script lang="ts">
import { ref } from 'vue'
import axios from 'axios'

export default {
  data() {
    return {
      location: null,
      videoId: null,
      videoUrl: null,
      title: null,
      thumbnailUrl: '',
      clipStart: 0
    }
  },
  methods: {
    getVideoInfo() {
      var url
      try {
        url = new URL(this.$data.videoUrl)
      } catch {
        return
      }

      let params = url.searchParams
      if (!params.has('v')) {
        return
      }

      if (params.has('t')) {
        this.$data.clipStart = Number(params.get('t'))
      }

      axios
        .post('http://localhost:7131/video', {
          type: this.$data.location,
          videoUrl: this.$data.videoUrl,
          save: false
        })
        .then((response) => {
          this.$data.title = response.data.title
          this.$data.thumbnailUrl = response.data.thumbnails[0].URL
          this.$data.videoId = response.data.id
        })
    },
    createVideo() {
      axios
        .post('http://localhost:7131/video', {
          type: this.$data.location,
          videoUrl: this.$data.videoUrl,
          save: true
          // clipStart: this.getSeconds(this.$data.clipStart),
          // clipEnd: this.getSeconds(this.$data.clipEnd)
        })
        .then((response) => {
          this.$router.push({
            name: 'video',
            params: {
              id: this.$data.videoId
            },
            query: {
              clipStart: this.$data.clipStart
            }
          })
          // TODO create session
          console.log('received response')
          console.log(response)
          // this.$data.preview = true
          // this.$data.id = response.data.id
          // this.$data.title = response.data.title
          // this.loadPreview(response);
        })
        .catch((err) => {
          console.debug(err)
        })
    }
  },
  mounted() {}
}
</script>
