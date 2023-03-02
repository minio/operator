import React from "react";
import { IIcon, selected, unSelected } from "./common";

const DescriptionIcon = ({ active = false }: IIcon) => {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="20"
      height="20"
      viewBox="0 0 24 24"
    >
      <path
        fill={active ? selected : unSelected}
        d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm5 11H7v-2h10v2z"
      ></path>
    </svg>
  );
};

export default DescriptionIcon;
