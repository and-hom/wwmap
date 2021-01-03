import React from 'react';

import ReactDOM from 'react-dom'
import {
    Route,
    Switch,
    Redirect,
    BrowserRouter
} from "react-router-dom";

import 'materialize-css';
import 'materialize-css/dist/css/materialize.min.css';

import {createBrowserHistory} from "history";

import Main from './main'
import River from './river'

class App extends React.Component {
    history = createBrowserHistory()
    render() {
        return (
            <div className="App">
                <Switch>
                    <Route history={history} path='/index' component={Main} />
                    <Route history={history} path='/river/:id' component={River} />
                    <Redirect from='/' to='/index'/>
                </Switch>
            </div>
        );
    }
}

ReactDOM.render((
    <BrowserRouter>
        <App/>
    </BrowserRouter>
), document.getElementById('root'))
