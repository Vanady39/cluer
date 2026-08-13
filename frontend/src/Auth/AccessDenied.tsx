import { userManager } from "./oidc";

export function AccessDenied() {
  const handleLogout = async () => {
    await userManager.signoutRedirect();
  };

  return (
    <main
      style={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
      }}
    >
      <div
        style={{
          textAlign: "center",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          gap: 16,
        }}
      >
        <h1 style={{ margin: 0 }}>Нет доступа</h1>

        <p style={{ margin: 0 }}>
          У этой учётной записи нет прав администратора.
        </p>

        <button
          type="button"
          onClick={handleLogout}
          style={{
            marginTop: 8,
            padding: "10px 20px",
            borderRadius: 8,
            border: "1px solid #ccc",
            background: "#fff",
            cursor: "pointer",
            fontSize: 14,
          }}
        >
          Выйти
        </button>
      </div>
    </main>
  );
}
