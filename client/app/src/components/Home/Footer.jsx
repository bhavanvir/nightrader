import LogoIcon from "../../assets/icons/LogoIcon";
import GitHubIcon from "../../assets/icons/GitHubIcon";

export default function Footer() {
  return (
    <footer className="footer bg-base-100 text-neutral-content items-center p-4">
      <aside className="grid-flow-col items-center">
        <LogoIcon />
        <p>Copyright Nightrader © 2024 - All right reserved</p>
      </aside>
      <nav className="grid-flow-col gap-4 md:place-self-center md:justify-self-end">
        <a href="https://github.com/bhavanvir/day-trader">
          <GitHubIcon />
        </a>
      </nav>
    </footer>
  );
}
