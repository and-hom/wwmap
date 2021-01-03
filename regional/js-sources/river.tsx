import React, {Component} from 'react'
import ReactMarkdown from 'react-markdown'
import {Link} from 'react-router-dom';
import {RiverFull} from "./data-model/model";
import {apiBase} from "./config";
import {LoadingState} from "./util";
import {Loading} from "./loading";


export default class River extends Component<{}, LoadingState<RiverFull>> {
    constructor(props) {
        super(props);
        this.state = {
            error: null,
            isLoaded: false,
            payload: null
        };
    }

    componentDidMount() {
        let id: bigint = this.props.match.params.id;
        fetch(`${apiBase}/river-card/${id}`)
            .then(res => res.json())
            .then(river => {
                    this.setState({
                        isLoaded: true,
                        payload: river,
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
        let river = this.state.payload
        return (
            <Loading loadingState={this.state}>
                <div className="container">
                    <div className="row">
                        <div className="col s9">
                            <h3>{river?.title}</h3>
                        </div>
                        <div className="col s3">
                            <Link to={'/index'}>Назад</Link>
                        </div>
                    </div>
                    <div className="row">
                        <div className="col s9">
                            <ReactMarkdown>{river?.description}</ReactMarkdown>
                        </div>
                        <div className="col s3">
                            <table className="wwmap-this.state.payload-download-table">
                                <tbody>
                                <tr>
                                    <td><label htmlFor="gpx">GPX для навигатора с русскими названиями</label></td>
                                    <td><a id="gpx" href={`${apiBase}/downloads/this.state.payload/${river?.id}/gpx`}
                                           title="Скачать GPX с точками порогов">GPX</a></td>
                                </tr>
                                <tr>
                                    <td><label htmlFor="gpx_en">GPX для навигатора без поддержки русских букв</label>
                                    </td>
                                    <td><a id="gpx_en" href={`${apiBase}/downloads/this.state.payload/${river?.id}/gpx?tr=true`}
                                           title="Скачать GPX с точками порогов">GPX<sub>en</sub></a></td>
                                </tr>
                                <tr>
                                    <td><label htmlFor="csv_en">Пороги таблицей</label></td>
                                    <td><a id="csv_en" href={`${apiBase}/downloads/this.state.payload/${river?.id}/csv`}
                                           title="Скачать таблицу с точками порогов">CSV</a></td>
                                </tr>
                                <tr>
                                    <td><label htmlFor="csv_en">Пороги таблицей латиницей</label></td>
                                    <td><a id="csv_en" href={`${apiBase}/downloads/this.state.payload/${river?.id}/csv?tr=true`}
                                           title="Скачать таблицу с точками порогов">CSV<sub>en</sub></a></td>
                                </tr>

                                </tbody>
                            </table>
                        </div>
                    </div>
                    <div className="row">
                        <div className="col s12">
                            <ul>
                                {river?.reports.flatMap(g => g.reports).map(report => (
                                    <li><a target="_blank" href={report.url} title={report.title}><img
                                        src={report.source_logo_url}/>{report.title}</a></li>
                                ))}
                            </ul>
                        </div>
                    </div>
                </div>
            </Loading>
        )
    }
}