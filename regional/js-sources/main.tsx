import React from 'react'


// import wwmapApi = require('../../api/js/map.v2.1')

export default class Main extends React.Component {
    // componentDidMount() {
    //     wwmapApi.initWWMapRegional("wwmap-container", 'AB', null, {
    //         catalogLinkType: "wwmap",
    //     })
    // }

    render() {
        return (
            <div>
                <div className="container main-container">
                    <div className="row">
                        <div className="col s12">
                            <h1>Путеводитель по рекам Грузии</h1>
                            <div id="wwmap-container" className="wwmap-container"></div>
                        </div>
                    </div>
                    <div className="row">
                        <div className="col s6">
                            <div className="container">
                                <div className="row">
                                    <div className="col s12">
                                        <h2>Аджария</h2>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="cell s12">
                                        <h3><a href="ajariskali.htm">Аджарискали</a></h3>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="cell s12">
                                        <h3>Мачахела</h3>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div className="col s6">
                            <div className="container">
                                <div className="row">
                                    <div className="col s12">
                                        <h2>Самегрело</h2>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="cell s12">
                                        <h3>Техури</h3>
                                    </div>
                                </div>
                                <div className="row">
                                    <div className="cell s12">
                                        <h3>Хобицкали</h3>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        )
    }
}