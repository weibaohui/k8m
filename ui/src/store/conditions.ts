import { create } from 'zustand';
import { fetcher } from '@/components/Amis/fetcher';

interface ConditionsStore {
    reverseConditions: string[];
    initialized: boolean;
    initReverseConditions: () => Promise<void>;
}

const useConditionsStore = create<ConditionsStore>((set) => ({
    reverseConditions: [],
    initialized: false,
    initReverseConditions: async () => {
        try {
            const response = await fetcher({
                url: '/params/condition/reverse/list',
                method: 'get'
            });
            if (Array.isArray(response.data?.data)) {
                set({ reverseConditions: response.data.data, initialized: true });
            }
        } catch (error) {
            console.error('获取反转条件列表失败:', error);
            set({ initialized: true }); // 即使失败也标记为已初始化
        }
    }
}));

export default useConditionsStore;