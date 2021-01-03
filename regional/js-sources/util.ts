export interface LoadingState<T> {
    error?: any
    isLoaded: boolean,
    payload?: T,
}
