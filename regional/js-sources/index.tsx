import React from 'react';

import ReactDOM from 'react-dom'
import {
    Route,
    Switch,
    Redirect,
    withRouter,
    BrowserRouter
} from "react-router-dom";


import {createBrowserHistory} from "history";

import Main from './main'
import River from './river'

class App extends React.Component {
    history = createBrowserHistory()
    render() {
        return (
            <div className="App">
                aaa
                <Switch>
                    <Route history={history} path='/index' component={Main} />
                    <Route history={history} path='/river' component={River} />
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
