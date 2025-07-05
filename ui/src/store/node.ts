import {Metadata} from "@/store/pod.ts";

export interface Node {
    metadata: Metadata;
    status?: {
        allocatable?: {
            cpu?: string;
            memory?: string;
            pods?: number
        };
        capacity?: {
            cpu?: string;
            memory?: string;
            pods?: number
        }
    };
}

