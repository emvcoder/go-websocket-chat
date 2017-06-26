$(document).ready(function() {
	var conn, connected;
	var command = $("#send");
	var log = $("#log");

	let height = 0

	connected = false;

	let roomChanger = id => {
		connected ? null : console.log("Connection problems");
		let data = {
			MessType: 'roomch',
			Message: id
		}
		conn.send(JSON.stringify(data))	
	}

	writeMessage = data => {
		height += 54;
		let add = `<div class="chat__message"><div class="author">${data.Author}</div><div class="message">${data.Message}</div></div>`;

		$("#chat")
			.append(add)
			.animate({ scrollTop: height }, "fast");
	}

	$("#send").keypress(e => e.which === 13 ? $("#button_send").click() : null);

	if (window["WebSocket"]) {
		conn = new WebSocket("ws://localhost:8080/ws");
		conn.onopen = function() {
			connected = true;
		};
		conn.onclose = function(evt) {
			console.log("Connection closed");
		};
		conn.onmessage = function(evt) {
			var data = JSON.parse(evt.data)
			switch (data.MessType) {
				case "cmd":
					$("#rooms").html("");
					for (item in data.Rooms) {
						$("#rooms").append(`<button id="room_${item}" class="btn">Room ${item}</button>`)
						$("#room_"+item).click(e => {
							roomChanger(e.target.id.substring(5))
						})
					}
					break
				case "msg":
					writeMessage(data)
					break
			}
		}
	} else {
		window.alert("Sorry, your device does not supported");
	}

	$("#button_send").click(function(){
		let data = {
			MessType: 'message',
			Message: command.val()
		}
		conn.send(JSON.stringify(data))
		command.val("");
	});

	$("#changer_nick").click(() => {
		let data = {
			MessType: 'cmd',
			Message: command.val()
		}
		conn.send(JSON.stringify(data))
		command.val("");
	})
});