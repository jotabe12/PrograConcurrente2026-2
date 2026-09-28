# `processed/` — Directorio de Datos Procesados

Este directorio contiene el CSV limpio generado a partir del archivo Parquet fuente.  
**El CSV no está incluido en el repositorio** — debe generarse localmente siguiendo las instrucciones de abajo.

---

## ¿Por qué no está el CSV en el repositorio?

El archivo `yellow_tripdata_2026-01.csv` pesa ~22 MB y contiene 1,002,542 filas.  
Está excluido mediante `.gitignore` (`/processed/*.csv`) para evitar los límites de tamaño de GitHub y mantener el historial de Git ágil.

El archivo Parquet **fuente** (`raw/yellow_tripdata_2026-01.parquet`, 64 MB) sí está en el repo y es el punto de partida.

---

## Cómo regenerar el CSV procesado

### Prerrequisitos

```powershell
pip install pandas pyarrow
```

### Ejecución

Desde la raíz del repositorio:

```powershell
jupyter nbconvert --to notebook --execute TrabajoParcial.ipynb --output TrabajoParcial_executed.ipynb
```

O bien, abrir `TrabajoParcial.ipynb` en Jupyter / VS Code y ejecutar todas las celdas en orden.

**Resultado:** se escribe `processed/yellow_tripdata_2026-01.csv`.

---

## Filtros de limpieza aplicados

La limpieza se realiza en `TrabajoParcial.ipynb` sobre el dataset crudo de NYC Yellow Taxi Trip Records (enero 2026).

| Paso | Criterio | Motivo |
|---|---|---|
| Selección de columnas | Se retienen solo las 10 columnas del modelo; se descartan `tip_amount`, `tolls_amount`, `extra`, `mta_tax`, `improvement_surcharge` | Evitar *data leakage* de métricas calculadas post-viaje |
| Filtro de distancia | `trip_distance > 0` | Eliminar viajes sin desplazamiento real |
| Filtro de tarifa | `fare_amount > 0` | Eliminar registros con tarifa nula o negativa |
| Filtro de pasajeros | `passenger_count > 0` | Eliminar registros sin pasajeros |
| Coherencia temporal | `tpep_dropoff_datetime > tpep_pickup_datetime` | Garantizar duración del viaje positiva |
| Columna derivada | Se agrega `trip_time_min` = duración del viaje en minutos | Feature utilizado por el modelo de regresión |

### Resultado

| Métrica | Valor |
|---|---|
| Registros crudos | 1,482,418 |
| Registros tras limpieza | **1,002,542** |
| Tasa de retención | ~67.6% |
| Variable objetivo | `fare_amount` (regresión) |

---

## Columnas del CSV resultante

```
tpep_pickup_datetime, tpep_dropoff_datetime, passenger_count,
trip_distance, trip_time_min, RatecodeID, PULocationID,
DOLocationID, payment_type, fare_amount
```

Estas columnas son exactamente las que el parser Go espera (`go-code/dataset.go`, variable `requiredCSVColumns`).
