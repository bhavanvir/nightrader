import LogoIcon from "../../assets/icons/LogoIcon";
import { Link } from "react-router-dom";

export default function SignInForm() {
  return (
    <div class="relative flex h-screen flex-col justify-center overflow-hidden">
      <div class="border-primary m-auto w-full rounded-md border p-6 shadow-md ring-2 ring-gray-800/50 lg:max-w-lg">
        <div className="flex justify-center">
          <Link to="/">
            <a href="/" className="btn btn-ghost text-xl">
              <LogoIcon />
              <span className="text-3xl font-semibold">Nightrader</span>
            </a>
          </Link>
        </div>
        <form class="space-y-4">
          <div>
            <label class="label">
              <span class="label-text text-base">Email</span>
            </label>
            <input
              type="text"
              placeholder="Email Address"
              class="input input-bordered w-full"
            />
          </div>
          <div>
            <label class="label">
              <span class="label-text text-base">Password</span>
            </label>
            <input
              type="password"
              placeholder="Enter Password"
              class="input input-bordered w-full"
            />
          </div>
          <a href="/" class="text-xs  hover:underline">
            Forget Password?
          </a>
          <div>
            <button class="btn btn-block btn-primary">Sign in</button>
          </div>
        </form>
      </div>
    </div>
  );
}
