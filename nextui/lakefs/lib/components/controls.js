import React from 'react';
import Form from "react-bootstrap/Form";
import Alert from "react-bootstrap/Alert";

const defaultDebounceMs = 300;

function debounce(func, wait, immediate) {
    let timeout;
    return function() {
        let context = this, args = arguments;
        let later = function() {
            timeout = null;
            if (!immediate) func.apply(context, args);
        };
        let callNow = immediate && !timeout;
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
        if (callNow) func.apply(context, args);
    };
}

export const DebouncedFormControl = React.forwardRef((props, ref) => {
    const onChange = debounce(props.onChange, (props.debounce !== undefined) ? props.debounce : defaultDebounceMs);
    return (<Form.Control ref={ref} {...{...props, onChange}}/>);
});

export const Loading = () => {
    return (
        <Alert variant={"info"}>Loading...</Alert>
    )
}

export const Error = (error) => {
    return (
        <Alert variant={"danger"}>{error.message}</Alert>
    )
}

export const ActionGroup = ({ children, orientation = "left" }) => {
    return (
        <div role="toolbar" className={`float-${orientation} mb-2 btn-toolbar`}>
            {children}
        </div>
    )
}

export const ActionsBar = ({ children }) => {
    return (
        <div className="actions-bar mt-3">
            {children}
        </div>
    )
}
