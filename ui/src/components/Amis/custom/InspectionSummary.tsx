import React from 'react';

interface InspectionSummaryComponentProps {
    data: any;
}

const InspectionSummaryComponent = React.forwardRef<HTMLSpanElement, InspectionSummaryComponentProps>(({ data }, _) => {


    return (
        <span >
            <div>xxxxx</div>
        </span>
    );
});

export default InspectionSummaryComponent;
