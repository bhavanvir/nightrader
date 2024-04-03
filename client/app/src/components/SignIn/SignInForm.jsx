import React, { useState, useContext } from "react";
import { Link } from "react-router-dom";
import axios from "axios";

import LogoIcon from "../../assets/icons/LogoIcon";
import AuthApi from "../../AuthApi";

export default function SignInForm() {
  const Auth = useContext(AuthApi);
  const [errorMessage, setErrorMessage] = useState("");

  const [formData, setFormData] = useState({
    username: "",
    password: "",
  });

  const [fieldValidity, setFieldValidity] = useState({
    username: true,
    password: true,
  });

  const [errorMessages, setErrorMessages] = useState({
    username: "",
    password: "",
  });

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData((prevData) => ({
      ...prevData,
      [name]: value,
    }));

    setErrorMessages((prevErrors) => ({
      ...prevErrors,
      [name]: "",
    }));
  };

  const validateFields = () => {
    const usernameRegex = /^[a-zA-Z0-9_]+$/;
    const passwordRegex = /^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d@$!%*#?&]{8,}$/;

    const newFieldValidity = {
      username: usernameRegex.test(formData.username.trim()),
      password: passwordRegex.test(formData.password.trim()),
    };

    setFieldValidity(newFieldValidity);

    const newErrorMessages = {
      username: newFieldValidity.username
        ? ""
        : "Username must contain only letters, numbers, and underscores",
      password: newFieldValidity.password
        ? ""
        : "Password must be at least 8 characters long and include at least one letter and one number",
    };

    setErrorMessages(newErrorMessages);

    return Object.values(newFieldValidity).every((isValid) => isValid);
  };

  const handleSubmit = (e) => {
    e.preventDefault();

    const isValid = validateFields();

    if (isValid) {
      axios
        .post(
          "http://localhost/authentication/login",
          {
            user_name: formData.username,
            password: formData.password,
          },
          {
            withCredentials: true,
          }
        )
        .then(function (response) {
          localStorage.setItem("token", response.data.data.token);
          Auth.setAuth(true);
        })
        .catch(function (error) {
          setErrorMessage(error.response.data.message);
        });
    }
  };

  return (
    <div className="relative flex h-screen flex-col justify-center overflow-hidden">
      {errorMessage && (
        <div role="alert" className="alert alert-error">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-6 w-6 shrink-0 stroke-current"
            fill="none"
            viewBox="0  0  24  24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="2"
              d="M10  14l2-2m0  0l2-2m-2  2l-2-2m2  2l2  2m7-2a9  9  0  11-18  0  9  9  0  0118  0z"
            />
          </svg>
          <span>{errorMessage}</span>
        </div>
      )}
      <div className="border-primary m-auto w-full rounded-md border p-6 shadow-md ring-2 ring-gray-800/50 lg:max-w-lg">
        <div className="flex justify-center">
          <Link to="/">
            <a href="/" className="btn btn-ghost text-xl">
              <LogoIcon />
              <span className="text-3xl font-semibold">Nightrader</span>
            </a>
          </Link>
        </div>
        <form className="space-y-4" onSubmit={handleSubmit}>
          <div>
            <label htmlFor="username" className="username">
              <span className="label-text text-base">Username</span>
            </label>
            <input
              type="text"
              id="username"
              name="username"
              placeholder="Username"
              className={`input input-bordered w-full ${
                fieldValidity.username ? "" : "border-red-500"
              }`}
              value={formData.username}
              onChange={handleInputChange}
            />
            {errorMessages.username && (
              <p className="text-sm text-red-500">{errorMessages.username}</p>
            )}
          </div>
          <div>
            <label htmlFor="password" className="password">
              <span className="label-text text-base">Password</span>
            </label>
            <input
              type="password"
              id="password"
              name="password"
              placeholder="Enter Password"
              className={`input input-bordered w-full ${
                fieldValidity.password ? "" : "border-red-500"
              }`}
              value={formData.password}
              onChange={handleInputChange}
            />
            {errorMessages.password && (
              <p className="text-sm text-red-500">{errorMessages.password}</p>
            )}
          </div>
          <a href="/" className="text-xs hover:underline">
            Forget Password?
          </a>
          <div>
            <button type="submit" className="btn btn-block btn-primary">
              Sign in
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
