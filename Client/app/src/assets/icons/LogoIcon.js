import React from "react";

const LogoIcon = (props) => {
  return (
    <svg
      viewBox="0 0 200 200"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className="h-8 w-8 stroke-current"
    >
      <g clipPath="url(#clip0_381_1024)">
        <circle cx="100" cy="100" r="88" strokeWidth="24" />
        <mask
          id="mask0_381_1024"
          style={{ maskType: "alpha" }}
          maskUnits="userSpaceOnUse"
          x="0"
          y="0"
          width="200"
          height="201"
        >
          <circle cx="100" cy="100" r="100" fill="black" />
        </mask>
        <g mask="url(#mask0_381_1024)">
          <circle cx="66.1763" cy="58.824" r="74.7647" strokeWidth="24" />
        </g>
      </g>
      <defs>
        <clipPath id="clip0_381_1024">
          <rect width="200" height="200" transform="translate(0 0.000488281)" />
        </clipPath>
      </defs>
    </svg>
  );
};

export default LogoIcon;
