import React from 'react'
import {CountryData} from "./country-settings/common";
import {country} from "./country-settings/ab";

type RegionMapProps = {
    country: CountryData
}

export default class RegionMap extends React.Component<RegionMapProps> {
    constructor(props: RegionMapProps) {
        super(props)
    }

    componentDidMount() {
        window['loadMapWhenDivIsReady'](country.countryCode);
    }

    render() {
        return (
            <div id="wwmap-container" className="wwmap-container"></div>
        )
    }
}