import {create} from 'zustand'
import {Schema} from 'amis'
import ajax from '@/utils/ajax'

interface Store {
    schema: Schema
    loading: boolean
    initPage: (path: string) => void
}

const useStore = create<Store>((set) => ({
    schema: {
        type: 'page'
    },
    loading: false,
    initPage(path) {
        set({loading: true})
        let page = path.slice(1);
        page = page + '.json';
        const url = `/public/pages/${page}`;
        ajax.get(url).then(res => {
            set({
                schema: res.data
            })
        }).finally(() => {
            set({loading: false})
        })
    }
}))

export default useStore
