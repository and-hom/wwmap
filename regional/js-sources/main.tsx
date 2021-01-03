import React from 'react'
import {Link} from 'react-router-dom';

import './style/main.css'
import {country} from './country-settings/ab'
import RegionMap from "./region-map";
import {apiBase} from "./config";
import {Region, RegionWithRivers} from './data-model/model'
import {LoadingState} from "./util";


export default class Main extends React.Component<{}, LoadingState<RegionWithRivers[]>> {
    constructor(props) {
        super(props);
        this.state = {
            error: null,
            isLoaded: false,
            payload: []
        };
    }

    enrichRegionWithRivers(region: Region): Promise<RegionWithRivers> {
        return fetch(`${apiBase}/region/${region.id}/river`)
            .then(res => res.json())
            .then(rivers => {
                return {
                    id: region.id,
                    title: region.title,
                    rivers: rivers
                } as RegionWithRivers
            })
    }

    enrichRegionsWithRivers(regions: Region[]): Promise<RegionWithRivers[]> {
        return Promise.all<RegionWithRivers>(regions.map((r: Region) => this.enrichRegionWithRivers(r)))
    }

    componentDidMount() {
        fetch(`${apiBase}/country/code/${country.countryCode}/region`)
            .then(res => res.json())
            .then(regions => this.enrichRegionsWithRivers(regions))
            .then(regionsWithRivers => {
                    this.setState({
                        isLoaded: true,
                        payload: regionsWithRivers,
                    });
                },
                (error) => {
                    this.setState({
                        isLoaded: true,
                        error: error,
                    });
                }
            );
    }

    render() {
        const {error, isLoaded, payload} = this.state;
        let chunk: RegionWithRivers[] = []
        let chunks: RegionWithRivers[][] = [];
        for (let i = 0; i < payload.length; i++) {
            if (i % 2 == 0) {
                chunk = [payload[i]]
            } else {
                chunk.push(payload[i])
                chunks.push(chunk);

            }
        }

        return (
            <div>
                <div className="container main-container">
                    <div className="row">
                        <div className="col s12">
                            <h1>Путеводитель по рекам {country.countryNamePrepositionalCase}</h1>
                            <RegionMap country={country}/>
                        </div>
                    </div>

                    {chunks.map(chunk => (
                        <div className="row">
                            {chunk.map(region => (
                                <div className="col s6">
                                    <div className="container">
                                        <div className="row">
                                            <div className="col s12">
                                                <h2>{region.title}</h2>
                                            </div>
                                        </div>
                                        {region.rivers.map(river => (
                                            <div className="row">
                                                <div className="cell s12">
                                                    <h3><Link to={`/river/${river.id}`}>{river.title}</Link></h3>
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            ))}
                        </div>
                    ))}
                </div>
            </div>
        )
    }
}