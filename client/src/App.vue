<script setup lang="ts">
import { onMounted, onUnmounted, reactive } from 'vue';

type Coords = { x: number; y: number };

type Client = { id: string; coords: Coords };

type WSMessage = {
	senderId: string;
	message: { coords: Coords } | { clients: Client[] };
	messageType: MessageTypes;
};

enum MessageTypes {
	Auth = 0,
	MouseMove = 1,
}

const socket = new WebSocket('ws://localhost:8080/ws');
const userCursors = reactive<Record<string, Coords>>({});

const clientId = (Math.random() * 100000).toFixed(0).toString();

onMounted(() => {
	socket.onopen = () => {
		console.log('Connected to WebSocket server');
		socket.send(
			JSON.stringify({
				senderId: clientId,
				messageType: MessageTypes.Auth,
				coords: { x: 0, y: 0 },
			}),
		);
	};

	socket.onmessage = (event) => {
		const data = JSON.parse(event.data) as WSMessage;
		if (data.messageType === MessageTypes.Auth) {
			data.message.clients.forEach((c) => {
				userCursors[c.id] = c.coords;
			});
		} else {
			userCursors[data.senderId] = data.message.coords;
		}
		console.log(data);
	};
});

const sendCursorCoords = (event: MouseEvent) => {
	const x = event.clientX;
	const y = event.clientY;
	userCursors[clientId] = { x, y };
	sendWSMessage(x, y);
};

const sendWSMessage = (x: number, y: number) => {
	const message: WSMessage = {
		senderId: clientId,
		messageType: MessageTypes.MouseMove,
		message: {
			coords: { x, y },
		},
	};
	socket.send(JSON.stringify(message));
};

document.addEventListener('mousemove', sendCursorCoords);

onUnmounted(() => {
	document.removeEventListener('mousemove', sendCursorCoords);
});
</script>

<template>
	<div>
		<div v-for="[id, coords] in Object.entries(userCursors)" :key="id" :style="{ position: 'absolute', left: coords?.x + 'px', top: coords?.y + 'px', color: 'red' }">
			<template v-if="id != clientId">
				<img src="/src/assets/coursor.png" style="height: 60px; width: 60px" />
			</template>
		</div>
	</div>
</template>

<style scoped>
.logo {
	height: 6em;
	padding: 1.5em;
	will-change: filter;
	transition: filter 300ms;
}
.logo:hover {
	filter: drop-shadow(0 0 2em #646cffaa);
}
.logo.vue:hover {
	filter: drop-shadow(0 0 2em #42b883aa);
}
</style>
