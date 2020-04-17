import React, { useState, useEffect } from 'react';
import { ApiContext } from '../contexts.mjs';
import {initApi} from '../api.mjs'

const ApiProvider = ({children}) => {
    const [api, setApi] = useState();

    useEffect(() => {
        setApi(initApi())
    }, [])

    if(!api) {
        return <div>Loading...</div>
    }

    return (
        <ApiContext.Provider value={api}>
            {children}
        </ApiContext.Provider>
    )
}

export default ApiProvider;