import sys
import numpy as np
from scipy.integrate import solve_ivp

#######################
#####  CONSTANTES  ####
#######################
days = 24 * 3600                    # segundos/día
G = 6.6742e-20                      # km^3/kg/s^2
rmoon = 1737                        # radio de la Luna (km)
rearth = 6378                       # radio de la Tierra (km)
r12 = 384400                        # distancia Tierra-Luna (km)

m1 = 5974e21                        # masa de la Tierra (kg)
m2 = 7348e19                        # masa de la Luna (kg)
M = m1 + m2
pi_1 = m1 / M
pi_2 = m2 / M

mu1 = 398600                        # parámetro gravitacional Tierra (km^3/s^2)
mu2 = 4903.02                       # parámetro gravitacional Luna (km^3/s^2)
mu = mu1 + mu2

C1 = -1.676 
C2 = -1.66490460
C3 = -1.58091856
C_12 = 0.5 * (C1 + C2)
C_13 = 0.5 * (C1 + C3)

W = np.sqrt(mu / r12**3)            # velocidad angular (rad/s)

x1 = -pi_2 * r12                    # posición x de la Tierra en el sistema rotante
x2 = pi_1 * r12                     # posición x de la Luna

L1 = 321710                             # L1 distance (km)

# Parámetros de propulsión y otros
n = 4
F = 0.000000450                     
T_val = F * n                       # empuje en kN (fase con empuje positivo)
m_motor = 1.875
m_cap = 2.0
tol = 1e-12

# -----------------------------
# Funciones de tasa (derivadas)
# -----------------------------
def rates(t, f):
    """Fase con empuje activo (primera trayectoria)."""
    x, y, vx, vy, m = f
    r1_val = np.linalg.norm([x + pi_2 * r12, y])
    r2_val = np.linalg.norm([x - pi_1 * r12, y])
    v_val = np.linalg.norm([vx, vy])
    ax = 2 * W * vy + W**2 * x - mu1 * (x - x1) / (r1_val**3) - mu2 * (x - x2) / (r2_val**3) + (T_val / m) * (vx / v_val)
    ay = -2 * W * vx + W**2 * y - (mu1/(r1_val**3) + mu2/(r2_val**3)) * y + (T_val / m) * (vy / v_val)
    g0 = 9.807e-3  # km/s² — matches Go's MassRate() exactly
    Isp = 1650
    mdot = -T_val / (g0 * Isp)
    return [vx, vy, ax, ay, mdot]

def rates0(t, f):
    """Fase de coasting (sin empuje)."""
    x, y, vx, vy, m = f
    r1_val = np.linalg.norm([x + pi_2 * r12, y])
    r2_val = np.linalg.norm([x - pi_1 * r12, y])
    ax = 2 * W * vy + W**2 * x - mu1 * (x - x1) / (r1_val**3) - mu2 * (x - x2) / (r2_val**3)
    ay = -2 * W * vx + W**2 * y - (mu1/(r1_val**3) + mu2/(r2_val**3)) * y
    return [vx, vy, ax, ay, 0]

def rates_1(t, f):
    """Fase de frenado (empuje negativo) para inserción lunar."""
    x, y, vx, vy, m = f
    r1_val = np.linalg.norm([x + pi_2 * r12, y])
    r2_val = np.linalg.norm([x - pi_1 * r12, y])
    v_val = np.linalg.norm([vx, vy])
    T_neg = -F * n  # empuje invertido (frena)
    ax = 2 * W * vy + W**2 * x - mu1 * (x - x1) / (r1_val**3) - mu2 * (x - x2) / (r2_val**3) + (T_neg / m) * (vx / v_val)
    ay = -2 * W * vx + W**2 * y - (mu1/(r1_val**3) + mu2/(r2_val**3)) * y + (T_neg / m) * (vy / v_val)
    g0 = 9.807e-3 
    Isp = 1650
    mdot = -abs(T_neg) / (g0 * Isp)
    return [vx, vy, ax, ay, mdot]

# -----------------------------
# Funciones de eventos
# -----------------------------

def jacobiC(t, y):
    """
    Evento: Se dispara cuando la “constante de Jacobi modificada”
    alcanza un valor umbral.
    """
    x_val, y_val, vx, vy, _ = y
    v_val = np.linalg.norm([vx, vy])
    r1_val = np.linalg.norm([x_val + pi_2 * r12, y_val])
    r2_val = np.linalg.norm([x_val - pi_1 * r12, y_val])
    return 0.5 * v_val**2 - 0.5 * W**2 * (x_val**2 + y_val**2) - mu1 / r1_val - mu2 / r2_val - (-1.63907788)

jacobiC.terminal = True
jacobiC.direction = 0


def lagrian1(t, y):
    """
    Evento: Se dispara cuando la distancia ajustada al centro terrestre
    alcanza el valor L1.
    """
    x_val, y_val, _, _, _ = y
    r1_val = np.linalg.norm([x_val + pi_2 * r12, y_val])
    return r1_val - L1

lagrian1.terminal = True
lagrian1.direction = 0


def jacobiC1(t, y):
    """
    Evento: Se dispara cuando la constante de Jacobi alcanza el valor C1
    durante la fase de frenado.
    """
    x_val, y_val, vx, vy, _ = y
    v_val = np.linalg.norm([vx, vy])
    r1_val = np.linalg.norm([x_val + pi_2 * r12, y_val])
    r2_val = np.linalg.norm([x_val - pi_1 * r12, y_val])
    return 0.5 * v_val**2 - 0.5 * W**2 * (x_val**2 + y_val**2) - mu1 / r1_val - mu2 / r2_val - C1

jacobiC1.terminal = True
jacobiC1.direction = 0


def print_state(phase, t, y):
    earthdist = np.hypot(y[0] - x1, y[1])
    print(f"phase={phase} t={t/days:6.2f}d earthdist={earthdist:.1f}km pos=({y[0]:.3f},{y[1]:.3f})")


def run(max_step, verify=False):
    h_apogee = 37000
    h_perigee = 1200
    r_apogee = rearth + h_apogee
    r_perigee = rearth + h_perigee
    e = (r_apogee - r_perigee) / (r_apogee + r_perigee)
    v0 = np.sqrt(mu1 * (1 - e) / r_apogee) - W * r_apogee
    phi = 0.7505211952744961 * np.pi / 180
    gamma = 0.0

    x0 = r_apogee * np.cos(phi) + x1
    y0 = r_apogee * np.sin(phi)
    vx0 = v0 * (np.sin(gamma) * np.cos(phi) - np.cos(gamma) * np.sin(phi))
    vy0 = v0 * (np.sin(gamma) * np.sin(phi) + np.cos(gamma) * np.cos(phi))
    f0 = [x0, y0, vx0, vy0, 12.0]

    if verify:
        print_state(0, 0.0, f0)

    sol1 = solve_ivp(
        rates,
        [0, days * 360],
        f0,
        method='RK45',
        events=jacobiC,
        rtol=1e-9,
        atol=tol,
        max_step=min(450, max_step),
    )
    if sol1.t_events[0].size == 0:
        raise RuntimeError('phase 1: jacobi threshold not crossed')
    t1 = sol1.t_events[0][0]
    f1 = sol1.y[:, -1]
    if verify:
        print_state(1, t1, f1)

    sol2 = solve_ivp(
        rates0,
        [t1, t1 + 260 * days],
        f1,
        method='RK45',
        events=lagrian1,
        rtol=1e-9,
        atol=tol,
        max_step=min(200, max_step),
    )
    if sol2.t_events[0].size == 0:
        raise RuntimeError('during phase 2: L1 distance not reached')
    t2 = sol2.t_events[0][0]
    f2 = sol2.y[:, -1]
    if verify:
        print_state(2, t2, f2)

    sol3 = solve_ivp(
        rates_1,
        [t2, t2 + 25 * days],
        f2,
        method='RK45',
        events=jacobiC1,
        rtol=1e-9,
        atol=tol,
        max_step=min(days, max_step),
    )
    if sol3.t_events[0].size == 0:
        raise RuntimeError('during phase 3: C1 distance not reached while braking')
    t3 = sol3.t_events[0][0]
    f3 = sol3.y[:, -1]
    if verify:
        print_state(3, t3, f3)

    sol4 = solve_ivp(
        rates0,
        [t3, t3 + 20 * days],
        f3,
        method='RK45',
        rtol=1e-9,
        atol=tol,
        max_step=min(100, max_step),
    )
    t4 = sol4.t[-1]
    f4 = sol4.y[:, -1]
    if verify:
        print_state(4, t4, f4)

    return f4


def main():
    if len(sys.argv) < 2:
        print('missing arg')
        sys.exit(1)

    try:
        max_step = float(sys.argv[1])
    except ValueError as err:
        print('bad arg:', err)
        sys.exit(1)

    verify = len(sys.argv) == 3 and sys.argv[2] == 'v'

    try:
        run(max_step, verify)
    except Exception as err:
        print(err)
        sys.exit(1)


if __name__ == '__main__':
    main()
