import {Parser} from 'expr-eval';
import {MenuItem} from '@/types/menu';

interface MenuVisibilityContext {
    userRole: string[];
    isGatewayAPISupported: boolean;
    isOpenKruiseSupported: boolean;
    isIstioSupported: boolean;
}

export const shouldShowMenuItem = (item: MenuItem, context: MenuVisibilityContext): boolean => {
    if (item.show === undefined || item.show === null) {
        return true;
    }

    if (typeof item.show === 'boolean') {
        return item.show;
    }

    if (typeof item.show === 'string') {
        try {
            const evalContext = {
                user: {role: 'user'},
            };

            const parser = new Parser();

            // 注入预定义的方法
            parser.functions.contains = function (str: string | string[], substr: string) {
                if (typeof str !== 'string' || typeof substr !== 'string') {
                    return false;
                }
                return str.includes(substr);
            };

            parser.functions.isGatewayAPISupported = function () {
                return context.isGatewayAPISupported;
            };

            parser.functions.isIstioSupported = function () {
                return context.isIstioSupported;
            };

            parser.functions.isOpenKruiseSupported = function () {
                return context.isOpenKruiseSupported;
            };

            parser.functions.isPlatformAdmin = function () {
                return context.userRole.includes('platform_admin');
            };

            parser.functions.isUserHasRole = function (role: string) {
                return context.userRole.includes(role);
            };

            const expr = parser.parse(item.show);
            const result = expr.evaluate(evalContext);

            return Boolean(result);
        } catch (error) {
            console.error('评估显示表达式错误:', error);
            return false;
        }
    }

    return true;
};