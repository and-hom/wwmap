import {frontendBase} from "./config";


export function initPresets() {
    let riverLabel = ymaps.templateLayoutFactory.createClass(`
    <div class="wwmap-bubble">
        <svg height="44" width="14">
            <circle r="11" cx="13" cy="12" fill="white" stroke="{{ properties.color }}" stroke-width="3px"/>
        </svg>
        <div class="wwmap-bubble-text" style="border-color: {{ properties.color }};">
        <img style="height: 14px;" src="${frontendBase}/img/invisible.png"/>&nbsp;&nbsp;{{ properties.iconContent }}</div>
        <svg height="44" width="14">
            <path d="M 11,12 A 30,85 0 0 1 0,44 L 2,20 L 11,12" fill="{{ properties.color }}" stroke="{{ properties.color }}"/>
            <circle r="11" cx="0" cy="12" fill="white" stroke="{{ properties.color }}" stroke-width="3px"/>
        </svg>
    </div>`)


    ymaps.option.presetStorage.add("wwmap#test", {
        hintLayout: null,
        iconSize: [66, 26],
        iconOffset: [-50, -50],
        iconShape: {
            type: 'Rectangle',
            coordinates: [[0, 0], [70, 32]]
        },
        pane: 'overlaps',
        iconLayout: riverLabel,
    })
}
