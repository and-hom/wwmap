import React from 'react';
import {LoadingState} from "./util";

export class LoadingProps {
    loadingState: LoadingState<any>
}

export class Loading extends React.Component<LoadingProps> {
    render() {

        if (!this.props.loadingState.isLoaded) {
            return (<div>Loading....</div>)
        }
        if (this.props.loadingState.error) {
            return (<div>Ошибка загрузки: {this.props.loadingState.error}</div>)
        }
        if (this.props.loadingState.payload==null) {
            return null;
        }

        return (
            <div>
                {this.props.children}
            </div>
        )
    }
}