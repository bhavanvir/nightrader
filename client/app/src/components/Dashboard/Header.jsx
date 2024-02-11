import React from "react";
import LogoIcon from "../../assets/icons/LogoIcon";
import CryptoJS from "crypto-js";
import Cookies from "js-cookie";
import AuthApi from "../../AuthApi";

export default function Header({ user }) {
  // Use the useContext hook inside the functional component
  const Auth = React.useContext(AuthApi);

  // Function to handle logout
  const handleLogout = () => {
    Auth.setAuth(false);
    Cookies.remove("session_token");
  };

  // Function to generate the Gravatar URL
  const getGravatarURL = (username) => {
    const hash = CryptoJS.MD5(username.trim().toLowerCase()).toString();
    return `https://www.gravatar.com/avatar/${hash}?d=identicon`;
  };

  const gravatarURL = getGravatarURL(user?.user_name);

  return (
    <div className="navbar bg-base-100">
      <div className="flex-1">
        <a href="/" className="btn btn-ghost text-xl">
          <LogoIcon />
          <span className="font-bold">Nightrader</span>
        </a>
      </div>
      <div className="flex-none">
        <div className="dropdown dropdown-end">
          <div
            tabIndex={0}
            role="button"
            className="btn btn-ghost btn-circle avatar"
          >
            <div className="w-10 rounded-full">
              <img alt="Tailwind CSS Navbar component" src={gravatarURL} />
            </div>
          </div>
          <ul
            tabIndex={0}
            className="menu menu-sm dropdown-content bg-base-100 rounded-box z-[1] mt-3 w-52 p-2 shadow"
          >
            <li>
              <a href="/" className="justify-between">
                Profile
                <span className="badge">New</span>
              </a>
            </li>
            <li>
              <a href="/">Settings</a>
            </li>
            <li>
              <button onClick={handleLogout}>Logout</button>
            </li>
          </ul>
        </div>
      </div>
    </div>
  );
}
