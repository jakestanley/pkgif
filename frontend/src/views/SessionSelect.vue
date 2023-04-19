<template>
    <h1>Select a session</h1>
    <button @click="newSession()">New session</button>
    <table>
        <thead>
            <th>Id</th>
            <th>Title</th>
            <th>URL</th>
            <th></th>
        </thead>
        <tbody>
            <tr v-for="(clip,i) in clips">
                <td>{{ clip.id }}</td>
                <td>{{ clip.title}}</td>
                <td>{{ clip.videoUrl }}</td>
                <td><button @click="resumeSession(i)">Resume</button></td>
            </tr>
        </tbody>
    </table>
    <h1>Start a new session</h1>
</template>

<script lang="ts">
import { ref } from 'vue'

import { useSessionStore } from '@/stores/session'
import router from '@/router'
import axios from 'axios'

export default {
    data () {
        return {
            clips: []
        }
    },
    methods: {
        newSession() {
            this.$router.push('/select-source')
        },
        resumeSession(i) {
            this.$router.push({
                name: 'clip',
                params: {
                    id: this.clips[i].id
                }
            })
            console.log("resuming session " + i)
        }
    },
    mounted () {
        this.clips = axios
            .get("http://localhost:7131/clip")
            .then(response => {
                this.clips = response.data
            })
    },
}
</script>