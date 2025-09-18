import { useState, useEffect } from 'react';
import { fetcher } from '@/components/Amis/fetcher';

interface UserRoleResponse {
    role: string;
    cluster: string;
    menu_data: string;
}

export const useUserRole = () => {
    const [userRole, setUserRole] = useState<string>('');
    const [menuData, setMenuData] = useState<string>('');

    useEffect(() => {
        const fetchUserRole = async () => {
            try {
                const response = await fetcher({
                    url: '/params/user/role',
                    method: 'get'
                });

                if (response.data && typeof response.data === 'object') {
                    const role = response.data.data as UserRoleResponse;
                    setUserRole(role.role);
                    const menuDataString = role.menu_data;
                    if (menuDataString && typeof menuDataString === 'string' && menuDataString.trim()) {
                        try {
                            // 验证是否为有效的JSON
                            JSON.parse(menuDataString);
                            setMenuData(menuDataString);
                        } catch (error) {
                            console.warn('Invalid menu_data JSON from server, using empty string:', error);
                            setMenuData('');
                        }
                    } else {
                        setMenuData('');
                    }

                    const originCluster = localStorage.getItem('cluster') || '';
                    if (originCluster === "" && role.cluster !== "") {
                        localStorage.setItem('cluster', role.cluster);
                    }
                }
            } catch (error) {
                console.error('Failed to fetch user role:', error);
            }
        };

        fetchUserRole();
    }, []);

    return { userRole, menuData };
};