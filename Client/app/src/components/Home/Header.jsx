export default function Header() {
  return (
    <div className="navbar bg-base-100">
      <div class="flex-1">
        <a href="/" className="btn btn-ghost text-xl">
          <svg
            viewBox="0 0 200 200"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
            className="h-8 w-8 stroke-current"
          >
            <g clip-path="url(#clip0_381_1024)">
              <circle cx="100" cy="100" r="88" stroke-width="24" />
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
                <circle
                  cx="66.1763"
                  cy="58.824"
                  r="74.7647"
                  stroke-width="24"
                />
              </g>
            </g>
            <defs>
              <clipPath id="clip0_381_1024">
                <rect
                  width="200"
                  height="200"
                  transform="translate(0 0.000488281)"
                />
              </clipPath>
            </defs>
          </svg>
          <span className="font-bold">Nightrader</span>
        </a>
      </div>
      <div className="flex-none ">
        <div className="dropdown dropdown-end">
          <div
            tabIndex={0}
            role="button"
            className="btn btn-ghost btn-circle avatar"
          >
            <div className="w-10 rounded-full">
              <img
                alt="Tailwind CSS Navbar component"
                src="https://daisyui.com/images/stock/photo-1534528741775-53994a69daeb.jpg"
              />
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
              <a href="/">Logout</a>
            </li>
          </ul>
        </div>
      </div>
    </div>
  );
}
