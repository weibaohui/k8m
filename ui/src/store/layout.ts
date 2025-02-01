import {create} from 'zustand'

interface Store {
    collapse: boolean
    updateField: (key: string, value: unknown) => void
}

const useStore = create<Store>((set) => ({
    collapse: false,
    updateField(key: string, value: unknown) {
        set({
            [key]: value
        })
    }
}))

export default useStore