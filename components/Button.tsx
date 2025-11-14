import React from "react";

const Button = ({
  children,
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement>) => {
  return (
    <button
      {...props}
      className="px-4 py-2 rounded-xl transition transform duration-200 active:scale-95 hover:brightness-80 outline-none focus:ring-0"
    >
      {children}
    </button>
  );
};

export default Button;
