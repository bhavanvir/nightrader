import React, { useState } from "react";
import { Link } from "react-router-dom";
import axios from "axios";

import LogoIcon from "../../assets/icons/LogoIcon";
export default function SignUpForm() {
  const [showSuccessAlert, setShowSuccessAlert] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");

  const [formData, setFormData] = useState({
    name: "",
    username: "",
    password: "",
  });

  const [fieldValidity, setFieldValidity] = useState({
    name: true,
    username: true,
    password: true,
  });

  const [errorMessages, setErrorMessages] = useState({
    name: "",
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
    const nameRegex = /^[a-zA-Z\s]+$/;
    const usernameRegex = /^[a-zA-Z0-9_]+$/;
    const passwordRegex = /^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d@$!%*#?&]{8,}$/;

    const newFieldValidity = {
      name: nameRegex.test(formData.name.trim()),
      username: usernameRegex.test(formData.username.trim()),
      password: passwordRegex.test(formData.password.trim()),
    };

    setFieldValidity(newFieldValidity);

    const newErrorMessages = {
      name: newFieldValidity.name ? "" : "Please enter a valid name",
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
          "http://localhost/authentication/register",
          {
            name: formData.name,
            user_name: formData.username,
            password: formData.password,
          },
          {
            withCredentials: true,
          }
        )
        .then(function (response) {
          if (response.data.success) {
            setShowSuccessAlert(true);
            setTimeout(() => setShowSuccessAlert(false), 5000);
          }
        })
        .catch(function (error) {
          setErrorMessage(error.response.data.message);
        });
    }
  };

  return (
    <div className="relative flex h-screen flex-col justify-center overflow-hidden">
      {showSuccessAlert && (
        <div role="alert" className="alert alert-success grid">
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
              d="M9  12l2  2  4-4m6  2a9  9  0  11-18  0  9  9  0  0118  0z"
            />
          </svg>
          <span>
            Your account has been successfully created! You can now sign in
          </span>
        </div>
      )}
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
            <label htmlFor="name" className="name">
              <span className="label-text text-base">Name</span>
            </label>
            <input
              type="text"
              id="name"
              name="name"
              placeholder="Name"
              className={`input input-bordered w-full ${
                fieldValidity.name ? "" : "border-red-500"
              }`}
              value={formData.name}
              onChange={handleInputChange}
            />
            {errorMessages.name && (
              <p className="text-sm text-red-500">{errorMessages.name}</p>
            )}
          </div>
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
          <div>
            <button type="submit" className="btn btn-block btn-primary">
              Sign up
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
