import LogoIcon from "../../assets/icons/LogoIcon";

export default function Header() {
  return (
    <div className="navbar bg-base-100">
      <div className="flex-1">
        <a href="/" className="btn btn-ghost text-xl">
          <LogoIcon />
          <span className="font-bold">Nightrader</span>
        </a>
      </div>
    </div>
  );
}
