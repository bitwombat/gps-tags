let map;

const home = { lat: -31.457605903965163, lng: 152.64217334946048 };
async function initMap() {
	const { Map } = await google.maps.importLibrary("maps");
	const { AdvancedMarkerElement } = await google.maps.importLibrary("marker");

	map = new Map(document.getElementById("map"), {
		center: home,
		zoom: 17,
		mapId: "HOME_MAP",
	});

	const marker = new AdvancedMarkerElement({
		map: map,
		position: home,
		title: "Home",
	});

}

initMap();
