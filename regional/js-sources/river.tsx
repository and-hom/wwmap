import React, {Component} from 'react'


export default class River extends Component {
    render() {
        return (
            <div className="container">
                <div className="row">
                    <div className="cell s12">
                        <h1>Аджарискали</h1>
                    </div>
                </div>
                <div className="row">
                    <div className="col s9">
                        <ul>
                            <li><a target="_blank" href="http://www.tlib.ru/doc.aspx?id=36632&amp;page=1"
                                   title=" #5432: р.Мал.Лаба = (от.Кор.Умпырский) = кор.Черноречье = р.Уруштен = р.Лаба = пос.Мостовский"><img
                                src="http://wwmap.ru/img/report_sources/tlib.png"/>&nbsp;<strong>1994&nbsp;</strong>
                                #5432: р.Мал.Лаба = (от.Кор.Умпырский) = кор.Черноречье = р.Уруштен = р.Лаба =
                                пос.Мостовский</a></li>
                            <li><a target="_blank" href="https://www.risk.ru/blog/209368"
                                   title="2016 ноябрь Кавказ - Большая Лаба и Белая - Клуб Новый Бродяга"><img
                                src="http://wwmap.ru/img/report_sources/riskru.png"/>&nbsp;2016 ноябрь Кавказ -
                                Большая
                                Лаба и Белая - Клуб Новый Бродяга</a></li>
                        </ul>
                    </div>
                    <div className="col s3">
                        <table className="wwmap-river-download-table">
                            <tbody>
                            <tr>
                                <td><label htmlFor="gpx">GPX для навигатора с русскими названиями</label></td>
                                <td><a id="gpx" href="https://wwmap.ru/api/downloads/river/32/gpx"
                                       alt="Скачать GPX с точками порогов">GPX</a></td>
                            </tr>
                            <tr>
                                <td><label htmlFor="gpx_en">GPX для навигатора без поддержки русских букв</label></td>
                                <td><a id="gpx_en" href="https://wwmap.ru/api/downloads/river/32/gpx?tr=true"
                                       alt="Скачать GPX с точками порогов">GPX<sub>en</sub></a></td>
                            </tr>
                            <tr>
                                <td><label htmlFor="csv_en">Пороги таблицей</label></td>
                                <td><a id="csv_en" href="https://wwmap.ru/api/downloads/river/32/csv"
                                       alt="Скачать таблицу с точками порогов">CSV</a></td>
                            </tr>
                            <tr>
                                <td><label htmlFor="csv_en">Пороги таблицей латиницей</label></td>
                                <td><a id="csv_en" href="https://wwmap.ru/api/downloads/river/32/csv?tr=true"
                                       alt="Скачать таблицу с точками порогов">CSV<sub>en</sub></a></td>
                            </tr>

                            </tbody>
                        </table>
                    </div>
                </div>
                <div className="row">
                    <div className="cell s12">
                        <h1>Аджарискали</h1>
                    </div>
                </div>
            </div>
        )
    }
}