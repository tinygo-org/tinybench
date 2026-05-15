import numpy as np
import matplotlib.pyplot as plt
from scipy.integrate import solve_ivp
import time
from matplotlib.animation import FuncAnimation, PillowWriter

plt.rcParams.update({'font.size': 8})

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

L1 = 321710                         # distancia L1 (km)

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
    g0 = 9.807
    Isp = 1650
    mdot = -T_val * 1000 / (g0 * Isp)
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
    g0 = 9.807
    Isp = 1650
    mdot = -abs(T_neg) * 1000 / (g0 * Isp)
    return [vx, vy, ax, ay, mdot]

def jacobi_potential(x, y):
    """
    Calcula el potencial de Jacobi (o energía efectiva) en cada punto (x,y),
    asumiendo velocidad cero (v = 0).
    """
    r1 = np.sqrt((x + pi_2 * r12)**2 + y**2)
    r2 = np.sqrt((x - pi_1 * r12)**2 + y**2)
    return -0.5 * W**2 * (x**2 + y**2) - mu1 / r1 - mu2 / r2

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
    a_array = np.linspace(C2, C_13, 20)
    a_threshold = a_array[11]
    a_threshold = -1.63907788
    jacobi_val = 0.5 * v_val**2 - 0.5 * W**2 * (x_val**2 + y_val**2) - mu1 / r1_val - mu2 / r2_val - a_threshold
    return jacobi_val
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

def circular(t, y):
    """
    Evento: Se dispara en el periselenio o aposelenio cuando el producto
    punto entre la posición relativa a la Luna y la velocidad es cero.
    """
    x_val, y_val, vx, vy, _ = y
    r2_vec = np.array([x_val - pi_1 * r12, y_val])
    v_vec = np.array([vx, vy])
    return np.dot(r2_vec, v_vec)
circular.terminal = True
circular.direction = 0

def collision_event(t, y):
    """
    Evento: Se dispara cuando la distancia al centro lunar es igual al radio lunar.
    """
    x_val, y_val, vx, vy, m = y
    d_moon = np.linalg.norm([x_val - x2, y_val])
    return d_moon - rmoon
collision_event.terminal = True
collision_event.direction = 0

def capture_event(t, y):
    """
    Evento: Se dispara cuando la energía relativa a la Luna se vuelve negativa.
    """
    x_val, y_val, vx, vy, m = y
    r_rel = np.linalg.norm([x_val - x2, y_val])
    speed = np.linalg.norm([vx, vy])
    E = 0.5 * speed**2 - mu2 / r_rel
    return E
capture_event.terminal = True
capture_event.direction = -1

# -----------------------------
# Función auxiliar para trazar círculos (Tierra, Luna)
# -----------------------------
def circle(xc, yc, radius, num_points=361):
    theta = np.deg2rad(np.linspace(0, 360, num_points))
    x = xc + radius * np.cos(theta)
    y = yc + radius * np.sin(theta)
    return x, y

# -----------------------------
# Función para animar la trayectoria en "chunks" particionados
# -----------------------------
def animar_trayectoria_dual(sols, t_threshold, chunk_interval_initial=2, chunk_interval_secondary=0.5,
                            nombre_archivo='trayectoria_dual.gif'):
    """
    Genera una animación en la que se arma el gráfico de la trayectoria de forma acumulada
    en dos tramos:
      - Desde el inicio hasta t_threshold usando un intervalo mayor (chunk_interval_initial, en días).
      - Desde t_threshold hasta el final usando un intervalo menor (chunk_interval_secondary, en días).
    t_threshold se define (por ejemplo) como el fin de la maniobra de empuje (sol1.t[-1]).
    """
    # Concatenar datos de todas las fases
    tiempos = np.concatenate([sol.t for sol in sols])
    x_total = np.concatenate([sol.y[0] for sol in sols])
    y_total = np.concatenate([sol.y[1] for sol in sols])
    
    # Convertir intervalos a segundos
    step_initial = chunk_interval_initial * days
    step_secondary = chunk_interval_secondary * days
    
    # Armar arreglo de tiempos para cada tramo
    frame_times_initial = np.arange(tiempos[0], t_threshold + step_initial, step_initial)
    frame_times_secondary = np.arange(t_threshold + step_secondary, tiempos[-1] + step_secondary, step_secondary)
    frame_times = np.concatenate((frame_times_initial, frame_times_secondary))
    
    # Configurar la figura y ejes
    fig, ax = plt.subplots(figsize=(8, 8))
    ax.set_xlim(-400000, 450000)
    ax.set_ylim(-325000, 325000)
    ax.set_xlabel("x [km]", fontsize=14)
    ax.set_ylabel("y [km]", fontsize=14)
    ax.set_aspect('equal', adjustable='box')  # ✅ Mantener relación de aspecto
    ax.grid(True)

    # Ajustar tamaño de ticks
    ax.tick_params(axis='both', labelsize=12)

    # Dibujar Tierra y Luna
    earth_x, earth_y = circle(x1, 0, rearth)
    moon_x, moon_y = circle(x2, 0, rmoon)
    ax.fill(earth_x, earth_y, 'b', alpha=0.9, label='Tierra')
    ax.fill(moon_x, moon_y, 'g', alpha=0.9, label='Luna')
    ax.legend(fontsize=12)

    # Elementos de la animación: línea de trayectoria y punto actual
    traj_line, = ax.plot([], [], 'r-', lw=2)
    current_point, = ax.plot([], [], 'ko', markersize=5)
    
    def init():
        traj_line.set_data([], [])
        current_point.set_data([], [])
        ax.set_title("Trayectoria Tierra-Luna - Día 0", fontsize=16)
        return traj_line, current_point
    
    def update(frame_index):
        t_frame = frame_times[frame_index]
        indices = np.where(tiempos <= t_frame)[0]
        traj_line.set_data(x_total[indices], y_total[indices])
        if indices.size > 0:
            current_point.set_data(x_total[indices[-1]], y_total[indices[-1]])
        
        # Calcular día actual
        day_count = int(np.floor(t_frame / days))
        ax.set_title(f"Trayectoria Tierra-Luna - Día {day_count}", fontsize=16)

        return traj_line, current_point

    # Ajustar márgenes sin romper aspect ratio
    fig.tight_layout()

    ani = FuncAnimation(fig, update, frames=len(frame_times), init_func=init, blit=True, repeat=False)
    ani.save(nombre_archivo, writer=PillowWriter(fps=24))
    print("GIF guardado como", nombre_archivo)



# -----------------------------
# Función principal de la simulación
# -----------------------------
def trayectoria():
    start_time = time.time()
    # Parámetros iniciales
    h_apogee = 37000    # Altitud del apogeo (km)
    h_perigee = 1200     # Altitud del perigeo (km)
    r_apogee = rearth + h_apogee  # Radio en el apogeo
    r_perigee = rearth + h_perigee  # Radio en el perigeo

    # Cálculo del semieje mayor y la excentricidad de la órbita elíptica
    a = (r_apogee + r_perigee) / 2
    e = (r_apogee - r_perigee) / (r_apogee + r_perigee)

    # Velocidad en el apogeo para la órbita GTO (usando la ecuación de vis-viva)
    v0 = np.sqrt(mu1 * (1 - e) / r_apogee) - W * r_apogee

    # Otros parámetros
    gamma = 0    # Ángulo de vuelo inicial (grados)
    t0 = 0
    tf = days * 360 * 4   # Tiempo máximo de integración (s)
    r0 = r_apogee         # Radio inicial igual al apogeo

    # Selección de ángulo
    phi = 0.7505211952744961  # en grados

    # Condiciones iniciales en el sistema (usando el apogeo como punto de partida)
    phi_rad = np.deg2rad(phi)
    gamma_rad = np.deg2rad(gamma)
    x0 = r0 * np.cos(phi_rad) + x1
    y0 = r0 * np.sin(phi_rad)
    vx0 = v0 * (np.sin(gamma_rad) * np.cos(phi_rad) - np.cos(gamma_rad) * np.sin(phi_rad))
    vy0 = v0 * (np.sin(gamma_rad) * np.sin(phi_rad) + np.cos(gamma_rad) * np.cos(phi_rad))
    m0_val = 12
    f0 = [x0, y0, vx0, vy0, m0_val]
    
    # Fase 1: Trayectoria con empuje
    sol1 = solve_ivp(rates, [t0, tf], f0, method='RK45', events=jacobiC,
                     rtol=1e-9, atol=tol, max_step=450)
    print("Fase 1 completada, tiempo de evento:", sol1.t_events, sol1.y.shape)
    f1_final = sol1.y[:, -1]

    # Fase 2: Coasting (motores apagados)
    t_phase2 = [sol1.t[-1], sol1.t[-1] + days * 650]
    sol2 = solve_ivp(rates0, t_phase2, f1_final, method='RK45', events=lagrian1,
                     rtol=1e-9, atol=tol, max_step=200)
    print("Fase 2 completada, tiempo de evento:", sol2.t_events, sol2.y.shape)
    if sol2.t_events[0].size == 0:
        print('No se logró llegar a L1')
        exito = False
        f2_final = sol2.y[:, -1]
        tiempo_total = sol2.t[-1]
        masa_final = sol2.y[4, -1]
        detalles = {
            'phi': phi,
            'tiempo_total': tiempo_total,
            'masa_final': masa_final,
            'exito': exito,
            'time_exec': time.time() - start_time
        }
        return detalles
    else:
        print('Se alcanzó L1 CORRECTAMENTE')
        f2_final = sol2.y[:, -1]

    # Fase 3: Frenado para inserción lunar
    t_phase3 = [sol2.t[-1], sol2.t[-1] + days * 25]
    sol3 = solve_ivp(rates_1, t_phase3, f2_final, method='RK45',
                     events=[jacobiC1], rtol=1e-9, atol=tol)
    print("Fase 3 completada, tiempo de evento:", sol3.t_events, sol3.y.shape)
    f3_final = sol3.y[:, -1]

    # Fase 4: Coasting post-frenado (sin empuje)
    t_phase4 = [sol3.t[-1], sol3.t[-1] + days * 20]
    sol4 = solve_ivp(rates0, t_phase4, f3_final, method='RK45',
                     rtol=1e-9, atol=tol, max_step=100)
    print("Fase 4 completada, tiempo final:", sol4.t[-1], sol4.y.shape)
    f4_final = sol4.y[:, -1]

    # --- Cálculo y gráfico de la energía de captura combinada ---
    t_phase2_vals = sol2.t
    x_phase2 = sol2.y[0]
    y_phase2 = sol2.y[1]
    vx_phase2 = sol2.y[2]
    vy_phase2 = sol2.y[3]
    
    t_split_phase2 = sol2.t[0] + 0.75 * (sol2.t[-1] - sol2.t[0])
    mask_phase2 = sol2.t >= t_split_phase2
    
    r_rel_phase2 = np.sqrt((x_phase2 - x2)**2 + (y_phase2)**2)
    E_capture_phase2 = 0.5 * (vx_phase2**2 + vy_phase2**2) - mu2 / r_rel_phase2

    t_phase3_vals = sol3.t
    x_phase3 = sol3.y[0]
    y_phase3 = sol3.y[1]
    vx_phase3 = sol3.y[2]
    vy_phase3 = sol3.y[3]
    r_rel_phase3 = np.sqrt((x_phase3 - x2)**2 + (y_phase3)**2)
    E_capture_phase3 = 0.5 * (vx_phase3**2 + vy_phase3**2) - mu2 / r_rel_phase3

    t_phase4_vals = sol4.t
    x_phase4 = sol4.y[0]
    y_phase4 = sol4.y[1]
    vx_phase4 = sol4.y[2]
    vy_phase4 = sol4.y[3]
    r_rel_phase4 = np.sqrt((x_phase4 - x2)**2 + (y_phase4)**2)
    E_capture_phase4 = 0.5 * (vx_phase4**2 + vy_phase4**2) - mu2 / r_rel_phase4

    plt.figure()
    plt.plot(t_phase2_vals[mask_phase2] / days, E_capture_phase2[mask_phase2], 'r-', 
             label="Energía (último ¼ de Fase 2)")
    plt.plot(t_phase3_vals / days, E_capture_phase3, 'b-', label="Energía (Fase 3)")
    plt.plot(t_phase4_vals / days, E_capture_phase4, 'm-', label="Energía (Fase 4)")
    plt.xlabel("Tiempo (días)")
    plt.ylabel("Energía de captura [km²/s²]")
    plt.title("Evolución de la Energía de Captura")
    plt.legend()
    plt.grid(True)
    plt.show()
    
    end_time = time.time()
    totalTime = end_time - start_time
    print('Tiempo que demoró -- ', totalTime)
    print('Masa final ', sol4.y[4,-1])
    
    # Gráfico de la trayectoria completa con fondo de potencial de Jacobi
    x_vals = np.linspace(-100000, 500000, 500)
    y_vals = np.linspace(-250000, 250000, 500)
    X, Y = np.meshgrid(x_vals, y_vals)
    Z = jacobi_potential(X, Y)

    plt.figure(figsize=(8, 8))
    plt.plot(sol1.y[0], sol1.y[1], 'g-', label="Fase 1 (Empuje)")
    plt.plot(sol2.y[0], sol2.y[1], 'r-', label="Fase 2 (Coasting)")
    plt.plot(sol3.y[0], sol3.y[1], 'b-', label="Fase 3 (Frenado)")
    plt.plot(sol4.y[0], sol4.y[1], 'm-', label="Fase 4 (Coasting post-frenado)")
    fontForLabel = 20
    earth_x, earth_y = circle(x1, 0, rearth)
    moon_x, moon_y = circle(x2, 0, rmoon)
    plt.fill(earth_x, earth_y, 'b', alpha=0.9, label='Earth')
    plt.fill(moon_x, moon_y, 'g', alpha=0.9, label='Moon')

    levels = np.linspace(-1.6649, -1.60, 4)
    contours = plt.contour(X, Y, Z, levels=levels, cmap='viridis')
    plt.xlabel("x [km]", fontsize=fontForLabel)
    plt.ylabel("y [km]", fontsize=fontForLabel)
    plt.title("Transferencia Tierra-Luna", fontsize=12)
    
    plt.legend()
    plt.grid(True)
    plt.xlim(-400000,450000)
    plt.ylim(-325000,325000)
    plt.show()
    
    # Usar como punto de corte el final de la fase 1 (maniobra de empuje)
    t_threshold = sol1.t[-1]
    # Generar animación dual: 2 días por frame hasta t_threshold, 0.5 días por frame después
    animar_trayectoria_dual([sol1, sol2, sol3, sol4], t_threshold, chunk_interval_initial=2, chunk_interval_secondary=0.25,
                             nombre_archivo='trayectoria_dual.gif')

    detalles = {
        'phi': phi,
        'tiempo_total': totalTime,
        'masa_final': sol4.y[4, -1],
        'exito': True,
        'time_exec': totalTime
    }
    return detalles

if __name__ == '__main__':
    trayectoria()
